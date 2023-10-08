package main

import (
	"fmt"
	pb "github.com/calvarado2004/vehicle-positions/proto"
)

func main() {

	tripId := "123"

	licensePlate := "LIC123"

	vehicleDescriptor := pb.VehicleDescriptor{
		LicensePlate: &licensePlate,
	}

	vehicle := &pb.VehiclePosition{
		Trip: &pb.TripDescriptor{
			TripId: &tripId,
		},
		Vehicle: &vehicleDescriptor,
	}

	fmt.Println(vehicle.Vehicle.LicensePlate)
}
