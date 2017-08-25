package collection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/autlamps/delay-backend-collection/models"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Conf stores the underlying clients (db, redis, rabbitmq) before they are turned into real services
type Conf struct {
	ApiKey string
	Db     *sql.DB
	Redis  redis.Client
}

// Env stores abstracted services for dealing with data
type Env struct {
}

// EnvFromConf returns an env
func EnvFromConf(conf Conf) Env {
	return Env{}
}

// Start contains our main loop for calling the realtime api and extracting info
func (e *Env) Start() error {

	// TODO: move this to a proper env variable and a correct location
	apiKey := ""

	url := fmt.Sprintf("http://api.at.govt.nz/v1/public/realtime/tripupdates?api_key=%v", apiKey)

	resp, err := http.Get(url)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to call realtime api")
	}

	if resp.StatusCode == 403 {
		logrus.WithField("resp", resp).Fatal("Incorrect api key used to call realtime api")
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var tu models.TUAPIResponse

	err = decoder.Decode(&tu)

	if err != nil {
		logrus.WithField("err", err).Error("Failed to decode json response")
	}

	return nil
}
