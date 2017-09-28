package output

import (
	"encoding/json"
	"time"
)

// Models for the json output used in the /delays endpoint

// OutTrip is the final output for an individual trip running abnormally
type OutTrip struct {
	TripID         string   `json:"trip_id"`
	RouteID        string   `json:"route_id"`
	RouteLongName  string   `json:"route_long_name"`
	RouteShortName string   `json:"route_short_name"`
	NextStop       NextStop `json:"next_stop"`
	VehicleID      string   `json:"vehicle_id"`
	VehicleType    int      `json:"vehicle_type"`
	Lat            float64  `json:"lat"`
	Lon            float64  `json:"lon"`
}

// NextStop is the information for the next stop of an abnormally running service
type NextStop struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Lat              float64   `json:"lat"`
	Lon              float64   `json:"lon"`
	ScheduledArrival time.Time `json:"scheduled_arrival"`
	Eta              time.Time `json:"eta"`
	Delay            int       `json:"delay"`
}

// Out is the final output of 1 run of the collection service, ready to be saved into redis
type Out struct {
	Count      int       `json:"count"`
	Trips      []OutTrip `json:"trips"`
	ExecName   string    `json:"exec_name"`
	Created    int64     `json:"created"`
	ValidUntil int64     `json:"valid_until"`
}

func (o *Out) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}
