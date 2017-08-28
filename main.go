package main

import (
	"fmt"
	"time"

	"github.com/autlamps/delay-backend-collection/collection"
)

func main() {
	// Shell main for dev purposes, this is not how it will look!
	t := time.Now()

	// TODO: move this to a proper env variable and a correct location
	apiKey := ""

	env := collection.Env{ApiKey: apiKey}
	env.Start()

	fmt.Printf("Runtime: %v", time.Now().Sub(t))
}
