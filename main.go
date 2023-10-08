package main

import (
	"fmt"
	pb "github.com/calvardo2004/vehicle-positions/proto"
)

func main() {
	vehicle := &pb.VehiclePosition{
		VehicleId: "V123",
		// ... set other fields
	}

	fmt.Println(vehicle)
}
