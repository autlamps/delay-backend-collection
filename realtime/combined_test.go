package realtime

import (
	"reflect"
	"testing"
	"time"
)

func TestNewCombTrip(t *testing.T) {

	curTime := time.Now()

	tests := []struct {
		n   string
		vl  VLEntity
		tu  TUEntity
		cb  CombEntity
		err error
	}{
		{
			n: "Should work",
			vl: VLEntity{
				"1234",
				false,
				VLVehicle{VLTrip{"1234", "1234", "1234", 1},
					VLVehicleID{"1234"},
					VLPosition{1.23, 2.14},
					-1.234},
			},
			tu: TUEntity{
				"1234",
				false,
				TUUpdate{
					TUTrip{"1234", "1234", 0},
					TUVehicle{"1234"},
					TUStopUpdate{1, "1234", 0, TUStopTimeEvent{100, curTime, ARRIVAl}},
					curTime},
			},
			cb: CombEntity{
				TUEntity{
					"1234",
					false,
					TUUpdate{
						TUTrip{"1234", "1234", 0},
						TUVehicle{"1234"},
						TUStopUpdate{1, "1234", 0, TUStopTimeEvent{100, curTime, ARRIVAl}},
						curTime}},
				CombPosition{1.23, 2.14},
			},
			err: nil,
		},
		{
			n: "Should produce error",
			vl: VLEntity{
				"1234",
				false,
				VLVehicle{VLTrip{"1234", "1234", "1234", 1},
					VLVehicleID{"1234"},
					VLPosition{1.23, 2.14},
					-1.234},
			},
			tu: TUEntity{
				"1234",
				false,
				TUUpdate{
					TUTrip{"1235", "1234", 0},
					TUVehicle{"1234"},
					TUStopUpdate{1, "1234", 0, TUStopTimeEvent{100, curTime, ARRIVAl}},
					curTime},
			},
			cb:  CombEntity{},
			err: ErrMismatchTripID,
		},
	}

	for _, test := range tests {
		cmb, err := NewCombTrip(test.tu, test.vl)

		if err != test.err {
			t.Errorf("%v - Err mismatch. Expected %v, got %v", test.n, test.err, err)
		}

		if !reflect.DeepEqual(cmb, test.cb) {
			t.Errorf("%v - Combined mismatch. Expected %v, got %v", test.n, test.cb, cmb)
		}

	}
}
