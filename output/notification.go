package output

import "encoding/json"

// Models to be passed to the notification service

type Notification struct {
	TripID     string
	StopTimeID string
	Delay      int
	Lat        float64
	Lon        float64
}

func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}
