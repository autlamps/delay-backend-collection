package models

import (
	"testing"
	"time"
)

func TestTUArrivalUnmarshal(t *testing.T) {

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
		arrival := &TUArrival{}

		err := arrival.UnmarshalJSON([]byte(test.in))

		if err != nil {
			t.Fatalf("Test failed with err: %v", err)
		}

		if !arrival.Time.Equal(test.expectedTime) {
			t.Fatalf("Unmarshal produces incorect time. Expected %v, got %v", test.expectedTime, arrival.Time)
		}

		if arrival.Delay != test.expectedDelay {
			t.Fatalf("Unmarshal produces incorect delay. Expected %v, got %v", test.expectedDelay, arrival.Delay)
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
			t.Fatalf("Test failed with err: %v", err)
		}

		if !update.Timestamp.Equal(test.expectedTime) {
			t.Fatalf("Unmarshal produces incorect time. Expected %v, got %v", test.expectedTime, update.Timestamp)
		}
	}
}
