package models

import "time"

type TripService interface {
	GetByGTFSID
}

type Trip struct {
	TripID         string
	RouteID        string
	RouteLongName  string
	RouteShortName string
	NextStop       NextStop
	Delay          int
	VehicleID      string
	Lat            float64
	Lon            float64
}

type NextStop struct {
	ID      string
	Lat     float64
	Lon     float64
	Arrival time.Time
}


func