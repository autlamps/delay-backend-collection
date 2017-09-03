package main

import (
	"fmt"
	"time"

	"github.com/autlamps/delay-backend-collection/collection"
	_ "github.com/lib/pq"
)

func main() {
	// Shell main for dev purposes, this is not how it will look!
	//runtime.GOMAXPROCS(runtime.NumCPU())

	t := time.Now()

	// TODO: move this to a proper env variable and a correct location
	apiKey := ""

	conf := collection.Conf{DBURL: "postgresql://postgres:mysecretpassword@172.17.0.2/postgres?sslmode=disable", ApiKey: apiKey, WorkerNo: 2, MQURL: "amqp://guest:guest@172.17.0.3:5672/"}
	env, err := collection.EnvFromConf(conf)

	if err != nil {
		panic(err)
	}

	err = env.Start()

	if err != nil {
		fmt.Println(err)
	}

	env.Done()

	fmt.Printf("Runtime: %v", time.Now().Sub(t))
}
