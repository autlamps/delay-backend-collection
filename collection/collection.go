package collection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/autlamps/delay-backend-collection/realtime"
	"github.com/autlamps/delay-backend-collection/static"
	"github.com/sirupsen/logrus"
)

// Conf stores the underlying clients (db, redis, rabbitmq) before they are turned into real services
type Conf struct {
	ApiKey   string
	WorkerNo int
	Db       *sql.DB
	//Redis  redis.Client
}

// Env stores abstracted services for dealing with data
type Env struct {
	ApiKey   string
	WorkerNo int
	Trips    static.TripStore
}

// EnvFromConf returns an env from a given conf
func EnvFromConf(conf Conf) Env {
	return Env{ApiKey: conf.ApiKey, Trips: static.TripServiceInit(conf.Db), WorkerNo: conf.WorkerNo}
}

// Start contains our main loop for calling the realtime api and extracting info
func (env *Env) Start() error {
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

	wc := make(chan realtime.CombEntity, 500)

	for i := 0; i < env.WorkerNo; i++ {
		go env.processEntity(wc)
	}

	for _, c := range cmb {
		wc <- c
	}

	close(wc)

	// Block until all entities on the channel have been processed
	var ok bool
	for !ok {
		_, ok = <-wc
	}

	return nil
}

// GetRealtimeTripUpdates calls the Trip Update api with the url and key from the env
func (env *Env) GetRealtimeTripUpdates(rc chan<- TripUpdateResult) {
	urlWithKey := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/tripupdates?api_key=%v", env.ApiKey)

	resp, err := http.Get(urlWithKey)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime trip updates api")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		logrus.WithField("resp", resp).Fatal("Incorrect api key used to call trip updates api")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}

	}

	var tu realtime.TUAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tu)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to decode trip update json response")
		rc <- TripUpdateResult{realtime.TUAPIResponse{}, err}
	}

	rc <- TripUpdateResult{tu, nil}
}

func (env *Env) GetRealtimeVehicleLocations(rc chan<- VehicleLocationResult) {
	urlWithKey := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/vehiclelocations?api_key=%v", env.ApiKey)

	resp, err := http.Get(urlWithKey)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime vehiclelocations api")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	if resp.StatusCode == 403 {
		// TODO: review whether or not this should halt execution
		logrus.WithField("resp", resp).Fatal("Incorrect api key used to call vehicle locations api")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	var vl realtime.VLAPIResponse

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&vl)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to decode vehicle location json response")
		rc <- VehicleLocationResult{realtime.VLAPIResponse{}, err}
	}

	rc <- VehicleLocationResult{vl, nil}
}

func (env *Env) processEntity(ec <-chan realtime.CombEntity) {
	for e := range ec {
		fmt.Println(e.Update.Trip.TripID)
	}
}
