package collection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/autlamps/delay-backend-collection/output"
	"github.com/autlamps/delay-backend-collection/realtime"
	"github.com/autlamps/delay-backend-collection/static"
	"github.com/sirupsen/logrus"
)

// Conf stores the underlying clients (db, redis, rabbitmq) before they are turned into real services
type Conf struct {
	ApiKey string
	Db     *sql.DB
	//Redis  redis.Client
}

// Env stores abstracted services for dealing with data
type Env struct {
	ApiKey string
	Trips  static.TripStore
}

// EnvFromConf returns an env from a given conf
func EnvFromConf(conf Conf) Env {
	return Env{ApiKey: conf.ApiKey, Trips: static.TripServiceInit(conf.Db)}
}

// Start contains our main loop for calling the realtime api and extracting info
// TODO: break this up into testable functions!
func (env *Env) Start() error {
	urlWithKey := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/tripupdates?api_key=%v", env.ApiKey)

	resp, err := http.Get(urlWithKey)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime tripupdates api")
		return err
	}

	if resp.StatusCode == 403 {
		logrus.WithField("resp", resp).Fatal("Incorrect api key used to call realtime api")
		return err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var tu realtime.TUAPIResponse

	err = decoder.Decode(&tu)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to decode trip update json response")
		return err
	}

	late, lateIDs := lateEntities(tu.Response.Entity)

	queryURL := env.createQueryURL("vehicleid", lateIDs)

	vehicleLocations := fmt.Sprintf("https://api.at.govt.nz/v1/public/realtime/vehiclelocations?%v", queryURL)

	resp1, err := http.Get(vehicleLocations)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime vehicle locations api")
		return err
	}

	defer resp1.Body.Close()
	decoder = json.NewDecoder(resp1.Body)
	var vl realtime.VLAPIResponse

	err = decoder.Decode(&vl)

	if err != nil {
		logrus.WithField("err", err).Errorf("Failed to decode vehicle location json response")
		return err
	}

	vlMap := make(map[string]realtime.VLEntity)

	// Create a map of all vehicle location entities with vehicle id as the key
	for _, e := range vl.Response.Entity {
		vlMap[e.Vehicle.Vehicle.ID] = e
	}

	interTrips := []output.InterTrip{}

	// Combine all our trip update entities and vehicle location entities into intermediate entities
	for _, e := range late {

		newInterTrip, err := output.NewInterTrip(e, vlMap[e.Update.Vehicle.ID])

		if err != nil {
			logrus.WithField("err", err).Errorf("Tried to create new InterTrip")
			return err
		}

		interTrips = append(interTrips, newInterTrip)
	}

	return nil
}

// lateEntities returns late entities and the vehicle ids of late entities
func lateEntities(tu []realtime.TUEntity) ([]realtime.TUEntity, []string) {

	lateE := []realtime.TUEntity{}
	lateS := []string{}

	for _, e := range tu {

		if !e.IsAbnormal() {
			continue
		}

		lateE = append(lateE, e)
		lateS = append(lateS, e.Update.Vehicle.ID)
	}

	return lateE, lateS
}

// createQueryURL creates a url get query string for a given key and values also including the api key
func (env *Env) createQueryURL(key string, value []string) string {

	vs := ""

	for i, v := range value {

		if i == len(value) { // If last value don't add "," to the end
			vs = vs + v
			continue
		}

		vs = vs + v + ","
	}

	v := url.Values{}

	v.Add(key, vs)
	v.Add("api_key", env.ApiKey)

	return v.Encode()
}
