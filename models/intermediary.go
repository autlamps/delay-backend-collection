package models

import "errors"

// Intermediary models are a combination of the raw entities from the api before context is added from the static data

var ErrMismatchVehicleID = errors.New("models - intermediary: cannot combine models with mismatched vehicle ids")

type InterTrip struct {
	TUEntity
	Position InterPosition
}

type InterPosition struct {
	Lat float64
	Lon float64
}

// NewInterTrip combines a TUEntity and VLEntity which represent the same trip
func NewInterTrip(tu TUEntity, vl VLEntity) (InterTrip, error) {

	if tu.Update.Vehicle.ID != vl.Vehicle.Vehicle.ID {
		return InterTrip{}, ErrMismatchVehicleID
	}

	pos := InterPosition{Lat: vl.Vehicle.Position.Lat, Lon: vl.Vehicle.Position.Long}

	return InterTrip{tu, pos}, nil
}
