package output

import (
	"errors"

	"github.com/autlamps/delay-backend-collection/realtime"
)

// Intermediary output are a combination of the raw entities from the api before context is added from the static data

var ErrMismatchVehicleID = errors.New("output - intermediary: cannot combine output with mismatched vehicle ids")

type InterTrip struct {
	realtime.TUEntity
	Position InterPosition
}

type InterPosition struct {
	Lat float64
	Lon float64
}

// NewInterTrip combines a TUEntity and VLEntity which represent the same trip
func NewInterTrip(tu realtime.TUEntity, vl realtime.VLEntity) (InterTrip, error) {

	if tu.Update.Vehicle.ID != vl.Vehicle.Vehicle.ID {
		return InterTrip{}, ErrMismatchVehicleID
	}

	pos := InterPosition{Lat: vl.Vehicle.Position.Lat, Lon: vl.Vehicle.Position.Long}

	return InterTrip{tu, pos}, nil
}
