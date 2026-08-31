package main

import (
	"sync"

	math "math/rand/v2"

	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
	"github.com/Sed-Miyuki/OmniRoute/shared/util"
	"github.com/mmcloughlin/geohash"
)

type Service struct{
	drivers []*driverInMap
	mu 		sync.RWMutex
}

type driverInMap struct{
	Driver *pb.Driver
}

func NewService() *Service{
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverID string,packageSlug string) (*pb.Driver,error){
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex:=math.IntN(len(PredefinedRoutes))
	randomRoute:=PredefinedRoutes[randomIndex]
	randomPlate:=GenerateRandomPlate()
	randomAvatar:=util.GetRandomAvatar(randomIndex)

	geohash:=geohash.Encode(randomRoute[0][0],randomRoute[0][1])

	driver:=&pb.Driver{
		Id: driverID,
		GeoHash: geohash,
		Location: &pb.Location{
			Latitude: randomRoute[0][0],
			Longitude: randomRoute[0][1],
		},
		Name: "Wali",
		PackageSlug: packageSlug,
		CarPlate: randomPlate,
		ProfilePicture: randomAvatar,
	}

	s.drivers = append(s.drivers, &driverInMap{
		Driver: driver,
	})

	return driver,nil
}

func (s *Service) UnRegisterDriver(driverID string){
	s.mu.Lock()
	defer s.mu.Unlock()

	for i,driver:=range s.drivers{
		if driver.Driver.Id == driverID {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
		}
	}
}

func (s *Service) FindAvailableDrivers(packageType string) []string{
	var matchingDrivers []string

	for _,driver:=range s.drivers{
		if driver.Driver.PackageSlug==packageType{
			matchingDrivers = append(matchingDrivers, driver.Driver.Id)
		}
	}

	if len(matchingDrivers)==0{
		return []string{}
	}

	return matchingDrivers
}