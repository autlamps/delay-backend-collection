package realtime

import (
	"errors"
)

var ErrMismatchTripID = errors.New("realtime/combined: cannot combine update with mismatched trip id")
var ErrMismatchUpdateLengths = errors.New("realtime/combined: cannot combine updates of different lengths")

// CombEntity is the combination of a TUEntity and VLEntity
type CombEntity struct {
	TUEntity
	Position CombPosition
}

type CombPosition struct {
	Lat float64
	Lon float64
}

// NewCombTrip combines a TUEntity and VLEntity which represent the same trip
func NewCombTrip(tu TUEntity, vl VLEntity) (CombEntity, error) {

	if tu.Update.Trip.TripID != vl.Vehicle.Trip.TripID {
		return CombEntity{}, ErrMismatchTripID
	}

	pos := CombPosition{Lat: vl.Vehicle.Position.Lat, Lon: vl.Vehicle.Position.Long}

	return CombEntity{tu, pos}, nil
}

// CombineTripUpdates combines matching TUEntity and VLEntity into a single object
func CombineTripUpdates(tue TUEntities, vle VLEntities) ([]CombEntity, error) {
	var ct []CombEntity

	if len(tue) != len(vle) {
		return []CombEntity{}, ErrMismatchUpdateLengths
	}

	vlm := vle.ToMap()

	for _, e := range tue {
		cmb, err := NewCombTrip(e, vlm[e.Update.Trip.TripID])

		if err != nil {
			// TODO: what should be done here?
			continue
		}

		ct = append(ct, cmb)
	}

	return ct, nil
}
