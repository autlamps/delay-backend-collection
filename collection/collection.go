package collection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/autlamps/delay-backend-collection/models"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Conf stores the underlying clients (db, redis, rabbitmq) before they are turned into real services
type Conf struct {
	ApiKey string
	Db     *sql.DB
	Redis  redis.Client
}

// Env stores abstracted services for dealing with data
type Env struct {
	ApiKey string
}

// EnvFromConf returns an env
func EnvFromConf(conf Conf) Env {
	return Env{}
}

// Start contains our main loop for calling the realtime api and extracting info
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
	var tu models.TUAPIResponse

	err = decoder.Decode(&tu)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to decode trip update json response")
		return err
	}

	_, lateIDs := lateEntities(tu.Response.Entity)

	queryURL := env.createQueryURL("vehicleid", lateIDs)

	vehicleLocations := fmt.Sprintf("https://api.at.govt.nz/v1/public/realtime/vehiclelocations?%v", queryURL)

	resp1, err := http.Get(vehicleLocations)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime vehicle locations api")
		return err
	}

	defer resp1.Body.Close()
	decoder = json.NewDecoder(resp1.Body)
	var vl models.VLAPIResponse

	err = decoder.Decode(&vl)

	if err != nil {
		logrus.WithField("err", err).Errorf("Failed to decode vehicle location json response")
		return err
	}

	return nil
}

// lateEntities returns late entities and the vehicle ids of late entities
func lateEntities(tu []models.TUEntity) ([]models.TUEntity, []string) {

	lateE := []models.TUEntity{}
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
