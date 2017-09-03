package collection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sync"

	"time"

	"database/sql"

	"github.com/autlamps/delay-backend-collection/notify"
	"github.com/autlamps/delay-backend-collection/output"
	"github.com/autlamps/delay-backend-collection/realtime"
	"github.com/autlamps/delay-backend-collection/static"
	log "github.com/sirupsen/logrus"
)

// Conf stores urls and paramaters for various services before being turned into an environment
type Conf struct {
	ApiKey   string
	WorkerNo int
	DBURL    string
	MQURL    string
	//RURL     string
}

// Env stores abstracted services for dealing with data
type Env struct {
	db           *sql.DB
	apikey       string
	WorkerNo     int
	Trips        static.TripStore
	StopTimes    static.StopTimeStore
	Routes       static.RouteStore
	Notification notify.Notifier
	Log          *log.Logger
}

// EnvFromConf returns an env from a given conf
func EnvFromConf(conf Conf) (Env, error) {
	l := log.New()

	db, err := sql.Open("postgres", conf.DBURL)

	if err != nil {
		l.WithField("err", err).Fatalf("Failed to open db connection")
	}

	if err := db.Ping(); err != nil {
		l.Fatal(err)
	}

	n, err := notify.InitService(conf.MQURL)

	if err != nil {
		return Env{}, err
	}

	return Env{
		db:           db,
		apikey:       conf.ApiKey,
		Trips:        static.TripServiceInit(db),
		StopTimes:    static.StopTimeServiceInit(db),
		Routes:       static.RouteServiceInit(db),
		WorkerNo:     conf.WorkerNo,
		Notification: n,
		Log:          l,
	}, nil
}

// Start contains our main loop for calling the realtime api and extracting info
func (env *Env) Start() error {
	defer func() {
		if r := recover(); r != nil {
			env.Log.WithField("r", r).Errorf("Recovered from panic!")
		}
	}()

	turc := make(chan TripUpdateResult)
	vlrc := make(chan VehicleLocationResult)

	go env.GetRealtimeTripUpdates(turc)
	go env.GetRealtimeVehicleLocations(vlrc)

	tur := <-turc
	close(turc)
	tu, err := tur.Unpack()

	if err != nil {
		return err
	}

	vlr := <-vlrc
	close(vlrc)
	vl, err := vlr.Unpack()

	if err != nil {
		return err
	}

	cmb, err := realtime.CombineTripUpdates(tu.Response.Entities, vl.Response.Entities)

	if err != nil {
		return err
	}

	wc := make(chan realtime.CombEntity, 1000)
	var wg sync.WaitGroup

	wg.Add(env.WorkerNo)

	// Create our workers
	for i := 0; i < env.WorkerNo; i++ {
		go env.processEntity(wc, &wg)
	}

	// Dispatch work
	for _, c := range cmb {
		wc <- c
	}

	close(wc)

	// Block until all entities on the channel have been processed
	wg.Wait()

	return nil
}

// Done should be called when you are done with execution in order to clean up connections
func (env *Env) Done() {
	env.Notification.Close()
	env.db.Close()
}

// GetRealtimeTripUpdates calls the Trip Update api with the url and key from the env
func (env *Env) GetRealtimeTripUpdates(rc chan<- TripUpdateResult) {
	urlWithKey := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/tripupdates?api_key=%v", env.apikey)

	resp, err := http.Get(urlWithKey)

	if err != nil {
		env.Log.WithField("err", err).Error("Failed to call realtime trip updates api")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		env.Log.WithField("resp", resp).Fatal("Incorrect api key used to call trip updates api")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}

	}

	var tu realtime.TUAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tu)

	if err != nil {
		env.Log.WithField("err", err).Error("Failed to decode trip update json response")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
	}

	rc <- TripUpdateResult{tu, nil}
}

func (env *Env) GetRealtimeVehicleLocations(rc chan<- VehicleLocationResult) {
	urlWithKey := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/vehiclelocations?api_key=%v", env.apikey)

	resp, err := http.Get(urlWithKey)

	if err != nil {
		env.Log.WithField("err", err).Error("Failed to call realtime vehiclelocations api")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		env.Log.WithField("resp", resp).Fatal("Incorrect api key used to call vehicle locations api")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	var vl realtime.VLAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&vl)

	if err != nil {
		env.Log.WithField("err", err).Error("Failed to decode vehicle location json response")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	rc <- VehicleLocationResult{vl, nil}
}

func (env *Env) processEntity(ec <-chan realtime.CombEntity, wg *sync.WaitGroup) {
	defer wg.Done()

	for e := range ec {

		// At this point, we don't care about normal entities, we just want to know whats running late or early
		if !e.IsAbnormal() {
			continue
		}

		st, err := env.Trips.GetTripByGTFSID(e.Update.Trip.TripID)

		if err != nil {
			env.Log.WithField("err", err).Errorf("Failed to get trip from database")
			continue
		}

		sr, err := env.Routes.GetRouteByID(st.RouteID)

		if err != nil {
			env.Log.WithFields(log.Fields{"err": err, "trip-id": st.ID}).Errorf("Failed to get route from database")
			continue
		}

		sts, err := env.StopTimes.GetStopsByTripID(st.ID)

		if err != nil {
			env.Log.WithFields(log.Fields{"err": err, "trip-id": st.ID}).Errorf("Failed to get stoptimes from database")
			continue
		}

		var nextSt static.StopTime

		if e.Update.StopUpdate.Event.Type == realtime.DEPARTURE {
			// We don't -1 here because we actually want the info for the next stop,
			// not the one the trip is departing from.

			// Some trips appear to be departing from their final stop which is causing line 214 to panic.
			// Currently _trying_ to capture one of these events in logging. We are recovering further up in the execution
			if e.Update.StopUpdate.StopSequence > len(sts)-1 {
				env.Log.WithFields(log.Fields{
					"sts":      sts,
					"stop_seq": e.Update.StopUpdate.StopSequence,
					"len(sts)": len(sts),
					"entity":   e,
				}).Errorf("Trying to access a stop sequence greater than the index for the number of stops")

				continue
			}

			nextSt = sts[e.Update.StopUpdate.StopSequence] // This line sometimes panics :(
		} else if e.Update.StopUpdate.Event.Type == realtime.ARRIVAl {
			// We do -1 here though to shift to index
			nextSt = sts[e.Update.StopUpdate.StopSequence-1]
		}

		ot := createOutputTrip(sr, st, nextSt, e)

		nt := output.Notification{
			TripID:     ot.TripID,
			StopTimeID: nextSt.TripID,
			Delay:      ot.NextStop.Delay,
			Lat:        ot.Lat,
			Lon:        ot.Lon,
		}

		ntjson, err := nt.ToJSON()

		if err != nil {
			env.Log.WithField("err", err).Errorf("Failed to marshal notification struct")
			continue
		}

		err = env.Notification.Send(ntjson)

		if err != nil {
			env.Log.WithField("err", err).Errorf("Failed to send notification")
			continue
		}

		// delays
	}
}

// createOutputTrip takes in a route, trip, stoptime and combined realtime entity and produces an output trip
func createOutputTrip(r static.Route, t static.Trip, nxtst static.StopTime, cmb realtime.CombEntity) output.OutTrip {
	next := output.NextStop{
		ID:               nxtst.StopInfo.ID,
		Name:             nxtst.StopInfo.Name,
		Lat:              nxtst.StopInfo.Lat,
		Lon:              nxtst.StopInfo.Lon,
		Delay:            cmb.Update.StopUpdate.Event.Delay,
		ScheduledArrival: nxtst.Arrival,
	}

	if cmb.Update.StopUpdate.Event.Type == realtime.DEPARTURE {
		// Estimated time to the next stop based on the current delay and the expected arrival at the next stop
		next.Eta = nxtst.Arrival.Add(time.Duration(cmb.Update.StopUpdate.Event.Delay) * time.Second)
	} else if cmb.Update.StopUpdate.Event.Type == realtime.ARRIVAl {
		// If the trip is arriving at the next stop then the gtfs realtime api already gives us an eta
		next.Eta = cmb.Update.StopUpdate.Event.Time
	}

	ot := output.OutTrip{
		TripID:         t.ID,
		RouteID:        t.RouteID,
		RouteLongName:  r.LongName,
		RouteShortName: r.ShortName,
		VehicleID:      cmb.TUEntity.Update.Vehicle.ID,
		Lat:            cmb.Position.Lat,
		Lon:            cmb.Position.Lon,
		NextStop:       next,
	}

	return ot
}
