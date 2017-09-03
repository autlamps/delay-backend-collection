package output

import (
	"time"
)

// Models for the json output used in the /delays endpoint

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

type NextStop struct {
	ID               string
	Name             string
	Lat              float64
	Lon              float64
	ScheduledArrival time.Time
	Eta              time.Time
	Delay            int
}
