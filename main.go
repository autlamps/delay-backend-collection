package main

import (
	"runtime"

	"flag"
	"os"

	"os/signal"
	"syscall"

	"fmt"

	"strconv"

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
	flag.StringVar(&dburl, "DATABASE_URL", "", "database url")
	flag.StringVar(&mqurl, "RABBITMQ_URL", "", "message queue url")
	flag.StringVar(&rdurl, "REDIS_URL", "", "redis url")
	flag.IntVar(&workerno, "WORKERS", 0, "number of workers")
	flag.Parse()

	if apikey == "" {
		apikey = os.Getenv("API_KEY")
	}

	if dburl == "" {
		dburl = os.Getenv("DATABASE_URL")
	}

	if rdurl == "" {
		rdurl = os.Getenv("REDIS_URL")
	}

	if mqurl == "" {
		mqurl = os.Getenv("RABBITMQ_URL")
	}

	if workerno == 0 {
		no, err := strconv.ParseInt(os.Getenv("WORKERS"), 10, 32)

		if err != nil {
			panic(err)
		}

		workerno = int(no) // ParseInt returns an int64 but workerno is an int because flag.IntVar requires an int
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
