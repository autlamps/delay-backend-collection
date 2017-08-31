package collection

import "github.com/autlamps/delay-backend-collection/realtime"

type TripUpdateResult struct {
	Result realtime.TUAPIResponse
	Err    error
}

func (tur *TripUpdateResult) Unpack() (realtime.TUAPIResponse, error) {
	return tur.Result, tur.Err
}

type VehicleLocationResult struct {
	Result realtime.VLAPIResponse
	Err    error
}

func (tur *VehicleLocationResult) Unpack() (realtime.VLAPIResponse, error) {
	return tur.Result, tur.Err
}
