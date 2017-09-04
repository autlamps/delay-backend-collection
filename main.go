package main

import (
	"runtime"

	"flag"
	"os"

	"os/signal"
	"syscall"

	"fmt"

	"github.com/autlamps/delay-backend-collection/collection"
	_ "github.com/lib/pq"
)

var mqurl string
var dburl string
var apikey string
var rdurl string
var workerno int

func init() {
	flag.StringVar(&apikey, "API_KEY", "", "AT api key")
	flag.StringVar(&dburl, "DB_URL", "", "database url")
	flag.StringVar(&mqurl, "MQ_URL", "", "message queue url")
	flag.StringVar(&rdurl, "RD_URL", "", "redis url")
	flag.IntVar(&workerno, "WORKERS", 2, "number of workers")
	flag.Parse()

	if dburl == "" {
		dburl = os.Getenv("DB_URL")
	}

	if rdurl == "" {
		rdurl = os.Getenv("RD_URL")
	}

	if mqurl == "" {
		mqurl = os.Getenv("MQ_URL")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	conf := collection.Conf{
		DBURL:    dburl,
		ApiKey:   apikey,
		WorkerNo: workerno,
		MQURL:    mqurl,
		RDURL:    rdurl,
	}

	env, err := collection.EnvFromConf(conf)

	if err != nil {
		panic(err)
	}

	// Exit channel used to signal env.Start that we want to stop executing after the current collection
	// is done
	ec := make(chan bool)

	// Our blocking channel. Start sends true down this once it is ready to exit
	fc := make(chan bool)

	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	// This func on receiving the syscall signal from the channel sends true down our exit channel.
	// This is done because Start takes a bool channel for it's exit channel and this seemed nicer/more reusable
	// than changing Start to take in a os.Signal channel
	go func() {
		<-sc
		fmt.Println("Exit signal recieved")
		ec <- true
	}()

	go env.Start(ec, fc)

	<-fc
}
