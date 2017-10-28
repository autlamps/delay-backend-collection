package realtime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTUStopTimeEvent_UnmarshalJSON(t *testing.T) {

	// TODO: this location needs more research, possibly shouldn't be hardcoded? I dunno
	loc, err := time.LoadLocation("Pacific/Auckland")

	if err != nil {
		t.Fatalf("Test failed with err: %v", err)
	}

	tests := []struct {
		in            string
		expectedTime  time.Time
		expectedDelay int
	}{
		{`{"delay": 100, "time": 1503627170}`, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc), 100},
		{`{"delay": 100, "time": 1503627170.999}`, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc), 100},
	}

	for _, test := range tests {
		arrival := &TUStopTimeEvent{}

		err := arrival.UnmarshalJSON([]byte(test.in))

		if err != nil {
			t.Fatalf("Test failed with err: %v", err)
		}

		if !arrival.Time.Equal(test.expectedTime) {
			t.Errorf("Unmarshal produces incorect time. Expected %v, got %v", test.expectedTime, arrival.Time)
		}

		if arrival.Delay != test.expectedDelay {
			t.Errorf("Unmarshal produces incorect delay. Expected %v, got %v", test.expectedDelay, arrival.Delay)
		}
	}
}

func TestTUUpdateUnmarshal(t *testing.T) {
	loc, err := time.LoadLocation("Pacific/Auckland")

	if err != nil {
		t.Fatalf("Test failed with err: %v", err)
	}

	tests := []struct {
		in           string
		expectedTime time.Time
	}{
		{`{"trip": null, "vehicle": null, "stop_time_update": null, "timestamp": 1503627170}`, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc)},
		{`{"trip": null, "vehicle": null, "stop_time_update": null, "timestamp": 1503627170.99}`, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc)},
	}

	for _, test := range tests {
		update := &TUUpdate{}

		err := update.UnmarshalJSON([]byte(test.in))

		if err != nil {
			t.Errorf("Test failed with err: %v", err)
		}

		if !update.Timestamp.Equal(test.expectedTime) {
			t.Errorf("Unmarshal produces incorect time. Expected %v, got %v", test.expectedTime, update.Timestamp)
		}
	}
}

func TestTUEntity_IsAbnormal(t *testing.T) {

	tests := []struct {
		in       string
		expected bool
	}{
		{`{ "id": "e2601758-977c-e56e-287a-96107e203ac0", "is_deleted": false, "trip_update": { "trip": { "trip_id": "13221091879-20170807091914_v56.25", "route_id": "22106-20170807091914_v56.25", "schedule_relationship": 0 }, "vehicle": { "id": "3075" }, "stop_time_update": { "stop_sequence": 29, "stop_id": "8242", "schedule_relationship": 0, "arrival": { "delay": -210, "time": 1503810877.772 } }, "timestamp": 1503810877.772 } }`, false},
		{`{ "id": "13b0524f-970d-df1f-10f5-b60fcd1bf48e", "is_deleted": false, "trip_update": { "trip": { "trip_id": "4915094780-20170807091914_v56.25", "route_id": "91506-20170807091914_v56.25", "schedule_relationship": 0 }, "vehicle": { "id": "5243" }, "stop_time_update": { "stop_sequence": 17, "stop_id": "4012", "schedule_relationship": 0, "arrival": { "delay": -433, "time": 1503810710.853 } }, "timestamp": 1503810710.853 } }`, true},
		{`{ "id": "4596d653-10f0-1301-90f4-a0d7055a895b", "is_deleted": false, "trip_update": { "trip": { "trip_id": "4973094078-20170807091914_v56.25", "route_id": "97301-20170807091914_v56.25", "schedule_relationship": 0 }, "vehicle": { "id": "5246" }, "stop_time_update": { "stop_sequence": 1, "stop_id": "4134", "schedule_relationship": 0, "arrival": { "delay": 2949, "time": 1503810849.5 } }, "timestamp": 1503810849.5 } }`, true},
	}

	for _, test := range tests {

		entity := &TUEntity{}

		err := json.Unmarshal([]byte(test.in), entity)

		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if test.expected != entity.IsAbnormal() {
			t.Errorf("Late differs. Epxected %v, got %v. Delay: %v", test.expected, entity.IsAbnormal(),
				entity.Update.StopUpdate.Event.Delay)
		}

	}
}

func TestTUEntity_IsCancelled(t *testing.T) {

	tests := []struct {
		in       string
		expected bool
	}{
		{`{ "id": "e2601758-977c-e56e-287a-96107e203ac0", "is_deleted": false, "trip_update": { "trip": { "trip_id": "13221091879-20170807091914_v56.25", "route_id": "22106-20170807091914_v56.25", "schedule_relationship": 0 }, "vehicle": { "id": "3075" }, "stop_time_update": { "stop_sequence": 29, "stop_id": "8242", "schedule_relationship": 0, "arrival": { "delay": -210, "time": 1503810877.772 } }, "timestamp": 1503810877.772 } }`, false},
		{`{ "id": "13b0524f-970d-df1f-10f5-b60fcd1bf48e", "is_deleted": false, "trip_update": { "trip": { "trip_id": "4915094780-20170807091914_v56.25", "route_id": "91506-20170807091914_v56.25", "schedule_relationship": 0 }, "vehicle": { "id": "5243" }, "stop_time_update": { "stop_sequence": 17, "stop_id": "4012", "schedule_relationship": 3, "arrival": { "delay": -433, "time": 1503810710.853 } }, "timestamp": 1503810710.853 } }`, true},
	}

	for _, test := range tests {

		entity := &TUEntity{}

		err := json.Unmarshal([]byte(test.in), entity)

		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if test.expected != entity.IsAbnormal() {
			t.Errorf("Canelled differs. Epxected %v, got %v. Schedule Relationship: %v", test.expected, entity.IsCancelled(),
				entity.Update.Trip.ScheduleRelationship)
		}
	}
}

func TestTUStopUpdate_UnmarshalJSON(t *testing.T) {
	loc, err := time.LoadLocation("Pacific/Auckland")

	if err != nil {
		t.Fatalf("Test failed with err: %v", err)
	}

	tests := []struct {
		in            string
		expectedType  TUStopTimeEventType
		expectedDelay int
		expectedTime  time.Time
	}{
		{`{ "stop_sequence": 29, "stop_id": "8242", "schedule_relationship": 0, "arrival": { "delay": -210, "time": 1503627170 } }`, ARRIVAl, -210, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc)},
		{`{ "stop_sequence": 29, "stop_id": "8242", "schedule_relationship": 0, "departure": { "delay": -433, "time": 1503627170 } }`, DEPARTURE, -433, time.Date(2017, time.August, 25, 14, 12, 50, 0, loc)},
	}

	for _, test := range tests {

		update := &TUStopUpdate{}

		err := update.UnmarshalJSON([]byte(test.in))

		if err != nil {
			t.Errorf("Test failed with err: %v", err)
		}

		if update.Event.Delay != test.expectedDelay {
			t.Errorf("Delay differs. Expected %v, got %v", test.expectedDelay, update.Event.Delay)
		}

		if !update.Event.Time.Equal(test.expectedTime) {
			t.Errorf("Time differs. Expected %v, got %v", test.expectedTime, update.Event.Time)
		}

		if update.Event.Type != test.expectedType {
			t.Errorf("Type differs. Expected %v, got %v", test.expectedType, update.Event.Type)
		}
	}

}
