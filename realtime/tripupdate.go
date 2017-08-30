package realtime

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
	ID      string   `json:"id"`
	Deleted bool     `json:"is_deleted"`
	Update  TUUpdate `json:"trip_update"`
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
	StopSequence         int    `json:"stop_sequence"`
	StopID               string `json:"stop_id"`
	ScheduleRelationship int    `json:"schedule_relationship"`
	Event                TUStopTimeEvent
}

// TUStopTimeEventType is whether the trip is arriving or departing the given stop
type TUStopTimeEventType uint8

const (
	ARRIVAl = iota
	DEPARTURE
)

// TUStopTimeEvent states the number of seconds the service is running behind and the the vehicles eta for the next stop
type TUStopTimeEvent struct {
	Delay int `json:"delay"`
	Time  time.Time
	Type  TUStopTimeEventType
}

// Custom unmarshal to convert float to Time in TUStopTimeEvent
func (a *TUStopTimeEvent) UnmarshalJSON(data []byte) error {
	type StopTimeEvent TUStopTimeEvent

	temp := &struct {
		FloatTime float64 `json:"time"`
		*StopTimeEvent
	}{
		StopTimeEvent: (*StopTimeEvent)(a),
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

// Custom unmarshal to convert either arrival or departure stoptimevent in json to TUStopTimeEvent
func (u *TUStopUpdate) UnmarshalJSON(data []byte) error {

	type StopUpdate TUStopUpdate

	temp := &struct {
		Departure TUStopTimeEvent `json:"departure"`
		Arrival   TUStopTimeEvent `json:"arrival"`
		*StopUpdate
	}{
		StopUpdate: (*StopUpdate)(u),
	}

	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	// Check to see if the StopTimeEvent is either departure of arrival
	if temp.Departure.Time.Second() != 0 {
		u.Event.Time = temp.Departure.Time
		u.Event.Delay = temp.Departure.Delay
		u.Event.Type = DEPARTURE
	}

	if temp.Arrival.Time.Second() != 0 {
		u.Event.Time = temp.Arrival.Time
		u.Event.Delay = temp.Arrival.Delay
		u.Event.Type = ARRIVAl
	}

	return nil
}

// SECONDS_ABNORMAL is the number of seconds a trip has to stray from schedule in order to be considered abnormal
const SECONDS_ABNORMAL = 240

// IsAbnormal returns true if the entity is running abnormally (i.e late or early), false if it is running to schedule
func (e *TUEntity) IsAbnormal() bool {
	return !(e.Update.StopUpdate.Event.Delay > -SECONDS_ABNORMAL && e.Update.StopUpdate.Event.Delay < SECONDS_ABNORMAL)
}
