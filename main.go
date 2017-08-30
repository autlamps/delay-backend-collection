package main

import (
	"fmt"
	"time"

	"database/sql"

	"github.com/autlamps/delay-backend-collection/collection"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Shell main for dev purposes, this is not how it will look!
	t := time.Now()

	// TODO: move this into a proper env var/flag!
	db, err := sql.Open("postgres", "postgresql://postgres:mysecretpassword@172.17.0.2/postgres?sslmode=disable")

	if err != nil {
		logrus.WithField("err", err).Fatalf("Failed to open db connection")
	}

	if err := db.Ping(); err != nil {
		logrus.Fatal(err)
	}

	// TODO: move this to a proper env variable and a correct location
	apiKey := ""

	conf := collection.Conf{Db: db, ApiKey: apiKey}
	env := collection.EnvFromConf(conf)
	env.Start()

	fmt.Printf("Runtime: %v", time.Now().Sub(t))
}
