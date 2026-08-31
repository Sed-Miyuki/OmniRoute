package messaging

import pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"

const (
	FindAvailableDriversQueue = "find_available_drivers"
)

type TripEventData struct {
	Trip *pb.Trip	`json:"trip"`
}
