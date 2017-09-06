package realtime

// Models for response from /realtime/tripupdates endpoint

type VLHeader struct {
	Version       string  `json:"gtfs_realtime_version"`
	Incrementally int     `json:"incrementally"`
	Timestamp     float64 `json:"timestamp"`
}

type VLTrip struct {
	TripID               string `json:"trip_id"`
	RouteID              string `json:"route_id"`
	StartTime            string `json:"start_time"`
	ScheduleRelationship int    `json:"schedule_relationship"`
}

type VLVehicleID struct {
	ID string `json:"id"`
}

type VLPosition struct {
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
	//Bearing int     `json:"bearing"`
}

type VLVehicle struct {
	Trip      VLTrip      `json:"trip"`
	Vehicle   VLVehicleID `json:"vehicle"`
	Position  VLPosition  `json:"position"`
	Timestamp float64     `json:"timestamp"`
}

type VLEntity struct {
	ID        string    `json:"id"`
	IsDeleted bool      `json:"is_deleted"`
	Vehicle   VLVehicle `json:"vehicle"`
}

type VLResponse struct {
	Header   VLHeader   `json:"header"`
	Entities VLEntities `json:"entity"`
}

type VLAPIResponse struct {
	Status   string     `json:"status"`
	Response VLResponse `json:"response"`
	Error    string     `json:"error"`
}

// VLEntities is simply a slice of VLEntity
type VLEntities []VLEntity

// Converts a slice of VLEntity into a map with trip id as the key
func (vle VLEntities) ToMap() map[string]VLEntity {
	m := make(map[string]VLEntity)

	for _, e := range vle {
		m[e.Vehicle.Trip.TripID] = e
	}

	return m
}
