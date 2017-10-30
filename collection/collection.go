package collection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sync"

	"time"

	"database/sql"

	"github.com/autlamps/delay-backend-collection/naming"
	"github.com/autlamps/delay-backend-collection/notify"
	"github.com/autlamps/delay-backend-collection/objstore"
	"github.com/autlamps/delay-backend-collection/output"
	"github.com/autlamps/delay-backend-collection/realtime"
	"github.com/autlamps/delay-backend-collection/static"
	log "github.com/sirupsen/logrus"
)

// Conf stores urls and parameters for various services before being turned into an environment
type Conf struct {
	ApiKey   string
	WorkerNo int
	DBURL    string
	MQURL    string
	RDURL    string
}

// Env stores abstracted services for dealing with data
type Env struct {
	db           *sql.DB
	apikey       string
	execname     string
	WorkerNo     int
	Trips        static.TripStore
	StopTimes    static.StopTimeStore
	Routes       static.RouteStore
	Notification notify.Notifier
	ObjStore     objstore.Store
	Log          *log.Logger
}

// EnvFromConf returns an env from a given conf
func EnvFromConf(conf Conf) (Env, error) {
	l := log.New()

	db, err := sql.Open("postgres", conf.DBURL)

	if err != nil {
		l.WithField("err", err).Fatalf("Failed to open db connection")
		return Env{}, fmt.Errorf("Failed to open db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		l.Fatal(err)
		return Env{}, fmt.Errorf("Failed to ping db connection: %v", err)
	}

	n, err := notify.InitService(conf.MQURL)

	if err != nil {
		return Env{}, err
	}

	o, err := objstore.InitService(conf.RDURL)

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
		ObjStore:     o,
		Log:          l,
	}, nil
}

// Start contains our main loop for calling run. To stop the loop pass a bool over exit, start will then send a bool
// over finished once completed.
func (env *Env) Start(exit <-chan bool, finished chan<- bool) {
	first := true

	defer func() {
		if r := recover(); r != nil {
			env.Log.WithField("r", r).Errorf("Recovered from panic!")
		}
	}()

	for {
		select {
		case <-exit:
			env.Log.Infof("Exiting Start Loop")
			finished <- true
			return
		default:
			if !first {
				// Sleep at the top so we exit sooner if we receive a value on exit
				time.Sleep(30 * time.Second)
			} else {
				first = false
			}

			start := time.Now()

			// We give a human readable name to every execution of Run, this allows us to track the source of events
			// and errors in our logs more easily. These names aren't necessarily unique but it doesn't really mater.
			env.execname = naming.GetRandomName()

			env.Log.Infof("%v - Starting collection", env.execname)
			err := env.Run()

			if err != nil {
				env.Log.WithField("err", err).Errorf("%v - Error occurred in run", env.execname)
			}

			env.Log.Infof("%v - Finished collection in %v", env.execname, time.Now().Sub(start))
		}
	}
}

// Run is the act of running one "collection" it calls the realtime apis, combines the data, adds context to the data,
// determines if a trip is running late, notifies the notification service of the trip running late and finally
// outputs a list of all trips running late.
func (env *Env) Run() error {
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

	oc := make(chan output.OutTrip, 1000)
	wc := make(chan realtime.CombEntity, 1000)
	var wg sync.WaitGroup

	wg.Add(env.WorkerNo)

	// Create our workers
	for i := 0; i < env.WorkerNo; i++ {
		go env.processEntity(wc, oc, &wg)
	}

	// Dispatch work
	for _, c := range cmb {
		wc <- c
	}

	close(wc)

	// Block until all entities on the channel have been processed
	wg.Wait()

	close(oc)

	out := output.Out{
		ExecName:   env.execname,
		Created:    time.Now().Unix(),
		ValidUntil: time.Now().Add(time.Duration(30) * time.Second).Unix(),
		Count:      0,
	}

	// Add all our processed ouput trips to our output struct
	for o := range oc {
		out.Trips = append(out.Trips, o)
		out.Count++
	}

	ojsn, err := out.ToJSON()

	if err != nil {
		return err
	}

	// Currently setting delays key to expire after 40 seconds
	err = env.ObjStore.Save("delays", ojsn, 120)

	if err != nil {
		return err
	}

	return nil
}

// Done should be called when you are done with execution in order to clean up connections
func (env *Env) Done() {
	env.ObjStore.Close()
	env.Notification.Close()
	env.db.Close()
}

// Custom http client with a timeout
var httpclient = http.Client{
	Timeout: 20 * time.Second,
}

// GetRealtimeTripUpdates calls the Trip Update api with the url and key from the env
func (env *Env) GetRealtimeTripUpdates(rc chan<- TripUpdateResult) {
	req, _ := http.NewRequest("GET", "http://api.at.govt.nz/v2/public/realtime/tripupdates", nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", env.apikey)

	resp, err := httpclient.Do(req)

	if err != nil {
		env.Log.WithField("err", err).Errorf("%v - Failed to call realtime trip updates api", env.execname)
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
		return
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		env.Log.WithField("resp", resp).Fatal("%v - Incorrect api key used to call trip updates api", env.execname)
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
		return
	}

	var tu realtime.TUAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tu)

	if err != nil {
		env.Log.WithField("err", err).Errorf("%v - Failed to decode trip update json response", env.execname)
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
		return
	}

	rc <- TripUpdateResult{tu, nil}
}

func (env *Env) GetRealtimeVehicleLocations(rc chan<- VehicleLocationResult) {
	req, _ := http.NewRequest("GET", "http://api.at.govt.nz/v2/public/realtime/vehiclelocations", nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", env.apikey)

	resp, err := httpclient.Do(req)

	if err != nil {
		env.Log.WithField("err", err).Errorf("%v - Failed to call realtime vehiclelocations api", env.execname)
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
		return
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		env.Log.WithField("resp", resp).Fatalf("%v - Incorrect api key used to call vehicle locations api", env.execname)
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
		return
	}

	var vl realtime.VLAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&vl)

	if err != nil {
		env.Log.WithField("err", err).Errorf("%v - Failed to decode vehicle location json response", env.execname)
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
		return
	}

	rc <- VehicleLocationResult{vl, nil}
}

func (env *Env) processEntity(ec <-chan realtime.CombEntity, oc chan<- output.OutTrip, wg *sync.WaitGroup) {
	defer wg.Done()

	for e := range ec {

		// At this point, we don't care about normal entities, we just want to know whats running late or early
		if !e.IsAbnormal() {
			continue
		}

		st, err := env.Trips.GetTripByGTFSID(e.Update.Trip.TripID)

		if err != nil {
			env.Log.WithField("err", err).Errorf("%v - Failed to get trip from database", env.execname)
			continue
		}

		sr, err := env.Routes.GetRouteByID(st.RouteID)

		if err != nil {
			env.Log.WithFields(log.Fields{"err": err, "trip-id": st.ID}).Errorf("%v - Failed to get route from database", env.execname)
			continue
		}

		sts, err := env.StopTimes.GetStopTimesByTripID(st.ID)

		if err != nil {
			env.Log.WithFields(log.Fields{"err": err, "trip-id": st.ID}).Errorf("%v - Failed to get stoptimes from database", env.execname)
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
				}).Errorf("%v - Trying to access a stop sequence greater than the index for the number of stops", env.execname)

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
			StopTimeID: nextSt.ID,
			Delay:      ot.NextStop.Delay,
			Lat:        ot.Lat,
			Lon:        ot.Lon,
			Route:      sr,
			Trip:       st,
			StopTimes:  sts,
		}

		ntjson, err := nt.ToJSON()

		if err != nil {
			env.Log.WithField("err", err).Errorf("%v - Failed to marshal notification struct", env.execname)
			continue
		}

		err = env.Notification.Send(ntjson)

		if err != nil {
			env.Log.WithField("err", err).Errorf("%v - Failed to send notification", env.execname)
			continue
		}

		// Send our final output trip to the output channel
		oc <- ot
	}
}

// createOutputTrip takes in a route, trip, stoptime and combined realtime entity and produces an output trip
func createOutputTrip(r static.Route, t static.Trip, nxtst static.StopTime, cmb realtime.CombEntity) output.OutTrip {
	next := output.NextStop{
		StopTimeID:       nxtst.ID,
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

	// Giant hack to fix our database having UTC timezones
	loc, _ := time.LoadLocation("Pacific/Auckland")

	nextETAString := next.Eta.Format("15:04:05")
	next.Eta, _ = time.ParseInLocation("15:04:05", nextETAString, loc)
	// End giant hack

	ot := output.OutTrip{
		TripID:         t.ID,
		RouteID:        t.RouteID,
		RouteLongName:  r.LongName,
		RouteShortName: r.ShortName,
		VehicleID:      cmb.TUEntity.Update.Vehicle.ID,
		VehicleType:    r.RouteType,
		Lat:            cmb.Position.Lat,
		Lon:            cmb.Position.Lon,
		NextStop:       next,
	}

	return ot
}
