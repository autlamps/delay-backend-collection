package models

import (
	"encoding/json"
	"math"
	"time"
)

// Models for response from /realtime/tripupdates endpoint

// TUAPIResponse is the full response from the api
type TUAPIResponse struct {
	Status   string     `json:"status"`
	Response TUResponse `json:"response"`
	Error    string     `json:"error"`
}

// TUResponse is the response section of the apis response
type TUResponse struct {
	Header TUHeader   `json:"header"`
	Entity []TUEntity `json:"entity"`
}

// TUHeader is the header section of the apis response
type TUHeader struct {
	Version       string  `json:"gtfs_realtime_version"`
	Incrementally int     `json:"incrementally"`
	Timestamp     float64 `json:"timestamp"`
}

// TUEntity contains information about an individual trip
type TUEntity struct {
	ID      string     `json:"id"`
	Deleted bool       `json:"is_deleted"`
	Update  []TUUpdate `json:"trip_update"`
}

// TUUpdate is specific information about the status of a current trip
type TUUpdate struct {
	Trip       TUTrip       `json:"trip"`
	Vehicle    TUVehicle    `json:"vehicle"`
	StopUpdate TUStopUpdate `json:"stop_time_update"`
	Timestamp  time.Time
}

// TUTrip contains the trip id and route id for the active trip
type TUTrip struct {
	TripID               string `json:"trip_id"`
	RouteID              string `json:"route_id"`
	ScheduleRelationship int    `json:"schedule_relationship"`
}

// TUVehicle contains the id of the vehicle currently traveling on a trip
type TUVehicle struct {
	ID string `json:"id"`
}

// TUStopUpdate contains information regarding the next stop on the trip
type TUStopUpdate struct {
	StopSequence         int       `json:"stop_sequence"`
	StopID               string    `json:"stop_id"`
	ScheduleRelationship int       `json:"schedule_relationship"`
	Arrival              TUArrival `json:"arrival"`
}

// TUArrival states the number of seconds the service is running behind and the the vehicles eta for the next stop
type TUArrival struct {
	Delay int `json:"delay"`
	Time  time.Time
}

// Custom unmarshal to convert float to Time in TUArrival
func (a *TUArrival) UnmarshalJSON(data []byte) error {
	type Arrival TUArrival

	temp := &struct {
		FloatTime float64 `json:"time"`
		*Arrival
	}{
		Arrival: (*Arrival)(a),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	sec, _ := math.Modf(temp.FloatTime) // We don't care about nano second accuracy so lets just drop it
	a.Time = time.Unix(int64(sec), 0)

	return nil
}

// Custom unmarshal to convert float to Time in TUUpdate
func (a *TUUpdate) UnmarshalJSON(data []byte) error {
	type Update TUUpdate

	temp := &struct {
		FloatTime float64 `json:"timestamp"`
		*Update
	}{
		Update: (*Update)(a),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	sec, _ := math.Modf(temp.FloatTime) // Again We don't care about nano second accuracy so lets just drop it
	a.Timestamp = time.Unix(int64(sec), 0)

	return nil
}
