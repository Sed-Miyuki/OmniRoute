package messaging

import (
	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
	pbd "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
)

const (
	FindAvailableDriversQueue = "find_available_drivers"
	DriverCmdTripRequestQueue = "driver_cmd_trip_request"
	DriverTripResponseQueue = "driver_trip_response"
	NotifyRiderNoDriversFoundQueue = "notify_rider_no_drivers_found"
	NotifyDriverAssignQueue = "notify_driver_assign"
)

type TripEventData struct {
	Trip *pb.Trip	`json:"trip"`
}

type DriverTripResponseData struct{
	Driver *pbd.Driver `json:"driver"`
	TripID string `json:"tripID"`
	RiderID string `json:"riderID"`
}
