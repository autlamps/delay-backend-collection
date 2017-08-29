package models

import "database/sql"

// Trip represents a trip as stored in the database
type Trip struct {
	ID        string
	RouteID   string
	ServiceID string
	GTFSID    string
	Headsign  string
}

// TripStore defines methods that a concrete trip service should implement
type TripStore interface {
	GetTripByGTFSID(id string) (Trip, error)
}

// TripService implements TripStore for psql
type TripService struct {
	DB *sql.DB
}

// TripServiceInit initializes and returns a TripService with a given sql db connector
func TripServiceInit(db *sql.DB) *TripService {
	return &TripService{DB: db}
}

// GetTripByGTFSID returns a trip with the given gtfs trip id or an error
func (ts *TripService) GetTripByGTFSID(id string) (Trip, error) {
	t := Trip{}

	row := ts.DB.QueryRow("SELECT * FROM trips WHERE gtfs_trip_id = $1", id)
	err := row.Scan(&t.ID, &t.RouteID, &t.ServiceID, &t.GTFSID, &t.Headsign)

	if err != nil {
		return t, err
	}

	return t, nil
}
