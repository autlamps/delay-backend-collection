package output

import (
	"encoding/json"
	"time"
)

// Models for the json output used in the /delays endpoint

// OutTrip is the final output for an individual trip running abnormally
type OutTrip struct {
	TripID         string
	RouteID        string
	RouteLongName  string
	RouteShortName string
	NextStop       NextStop
	VehicleID      string
	Lat            float64
	Lon            float64
}

// NextStop is the information for the next stop of an abnormally running service
type NextStop struct {
	ID               string
	Name             string
	Lat              float64
	Lon              float64
	ScheduledArrival time.Time
	Eta              time.Time
	Delay            int
}

// Out is the final output of 1 run of the collection service, ready to be saved into redis
type Out struct {
	Count      int
	Trips      []OutTrip
	ExecName   string
	Created    int64
	ValidUntil int64
}

func (o *Out) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}
