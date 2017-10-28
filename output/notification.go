package output

import (
	"encoding/json"

	"github.com/autlamps/delay-backend-collection/static"
)

// Models to be passed to the notification service

type Notification struct {
	Cancelled  bool
	TripID     string
	StopTimeID string
	Delay      int
	Lat        float64
	Lon        float64
	Route      static.Route
	Trip       static.Trip
	StopTimes  []static.StopTime
}

func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}
