package main

import pb "github.com/calvarado2004/vehicle-positions/proto"

type BusPosition struct {
	ID        string
	Latitude  float64
	Longitude float64
	Label     string
	Bearing   float64
}

type VehiclePosition struct {
	Trip                *pb.TripDescriptor
	Vehicle             *pb.VehicleDescriptor
	Position            *pb.Position
	CurrentStopSequence *uint32
	StopId              *string
	CurrentStatus       *pb.VehiclePosition_VehicleStopStatus
	Timestamp           *uint64
	CongestionLevel     *pb.VehiclePosition_CongestionLevel
	OccupancyStatus     *pb.VehiclePosition_OccupancyStatus
}

type TripUpdate struct {
	Trip           *pb.TripDescriptor
	Vehicle        *pb.VehicleDescriptor
	StopTimeUpdate []*pb.TripUpdate_StopTimeUpdate
	Timestamp      *uint64
	Delay          *int32
}

type Route struct {
	ID        string `csv:"route_id"`
	ShortName string `csv:"route_short_name"`
	LongName  string `csv:"route_long_name"`
	Color     string `csv:"route_color"`
	TextColor string `csv:"route_text_color"`
	// Add other fields as required.
}

type Shape struct {
	ShapeId      string  `csv:"shape_id"`
	Latitude     float64 `csv:"shape_pt_lat"`
	Longitude    float64 `csv:"shape_pt_lon"`
	Sequence     int     `csv:"shape_pt_sequence"`
	DistTraveled float64 `csv:"shape_dist_traveled"`
}

type Stop struct {
	StopID    string  `csv:"stop_id"`
	StopCode  string  `csv:"stop_code"`
	StopName  string  `csv:"stop_name"`
	StopDesc  string  `csv:"stop_desc"`
	Latitude  float64 `csv:"stop_lat"`
	Longitude float64 `csv:"stop_lon"`
}

var currentBusPositions []BusPosition

type RouteVisualization struct {
	RouteInfo   Route
	Shapes      []Shape
	Stops       []Stop
	Buses       []BusVisualization
	TripUpdates []TripUpdate
}

type BusVisualization struct {
	BusPosition   BusPosition
	TripInfo      *pb.TripDescriptor
	StopSequences []*pb.TripUpdate_StopTimeUpdate
}
