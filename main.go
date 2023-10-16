package main

import (
	"encoding/csv"
	"encoding/json"
	pb "github.com/calvarado2004/vehicle-positions/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "status"})

	busCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bus_count",
			Help: "Total number of buses fetched from the API.",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(busCount)

}

func routeVisualizationHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	routeID := r.URL.Query().Get("route_id")
	if routeID == "" {
		http.Error(w, "Route ID not provided", http.StatusBadRequest)
		return
	}

	// Fetch all the necessary data
	routes, _ := ParseRoutes("./google_transit/routes.txt")
	var selectedRoute Route
	for _, route := range routes {
		if route.ID == routeID {
			selectedRoute = route
			break
		}
	}

	shapes, _ := ParseShapes("./google_transit/shapes.txt")

	stops, _ := ParseStops("./google_transit/stops.txt")

	martaBusPositionsURL := "https://gtfs-rt.itsmarta.com/TMGTFSRealTimeWebService/vehicle/vehiclepositions.pb"

	buses := getBusPositions(martaBusPositionsURL)
	busCount.Set(float64(len(buses)))

	martaTripUpdatesURL := "https://gtfs-rt.itsmarta.com/TMGTFSRealTimeWebService/tripupdate/tripupdates.pb"
	tripUpdates := getTripUpdates(martaTripUpdatesURL)

	var busVisualizations []BusVisualization
	for _, bus := range buses {
		for _, tripUpdate := range tripUpdates {
			if bus.ID == tripUpdate.Vehicle.GetId() {
				busVis := BusVisualization{
					BusPosition:   bus,
					TripInfo:      tripUpdate.Trip,
					StopSequences: tripUpdate.StopTimeUpdate,
				}
				busVisualizations = append(busVisualizations, busVis)
				break
			}
		}
	}

	routeVis := RouteVisualization{
		RouteInfo:   selectedRoute,
		Shapes:      shapes,
		Stops:       stops,
		Buses:       busVisualizations,
		TripUpdates: tripUpdates,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(routeVis)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
	duration := time.Since(start).Seconds()
	httpRequestDuration.WithLabelValues("/route-visualization").Observe(duration)
	httpRequestsTotal.WithLabelValues("/route-visualization", strconv.Itoa(http.StatusOK)).Inc()
}

func busPositionsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(currentBusPositions)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func tripUpdatesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	martaTripUpdatesURL := "https://gtfs-rt.itsmarta.com/TMGTFSRealTimeWebService/tripupdate/tripupdates.pb"

	err := json.NewEncoder(w).Encode(getTripUpdates(martaTripUpdatesURL))
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func shapesHandler(w http.ResponseWriter, r *http.Request) {

	shapes, err := ParseShapes("./google_transit/shapes.txt")
	if err != nil {
		log.Fatalf("Failed to parse shapes: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(shapes)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}

}

func routesHandler(w http.ResponseWriter, r *http.Request) {

	routes, err := ParseRoutes("./google_transit/routes.txt")
	if err != nil {
		log.Fatalf("Failed to parse routes: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(routes)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func stopsHandler(w http.ResponseWriter, r *http.Request) {
	stops, err := ParseStops("./google_transit/stops.txt")
	if err != nil {
		log.Fatalf("Failed to parse stops: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stops)
	if err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func main() {

	martaBusPositionsURL := "https://gtfs-rt.itsmarta.com/TMGTFSRealTimeWebService/vehicle/vehiclepositions.pb"

	currentBusPositions = getBusPositions(martaBusPositionsURL)
	busCount.Set(float64(len(currentBusPositions)))

	// Start fetching bus positions every 15 seconds
	go func() {
		for range time.Tick(1 * time.Second * 15) {
			currentBusPositions = getBusPositions(martaBusPositionsURL)
			busCount.Set(float64(len(currentBusPositions)))
			log.Println("Updated bus positions!")
		}
	}()

	handler := http.NewServeMux()
	handler.HandleFunc("/shapes", shapesHandler)
	handler.HandleFunc("/routes", routesHandler)
	handler.HandleFunc("/trip-updates", tripUpdatesHandler)
	handler.HandleFunc("/bus-positions", busPositionsHandler)
	handler.HandleFunc("/stops", stopsHandler)
	handler.HandleFunc("/route-visualization", routeVisualizationHandler)
	handler.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	handler.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "DELETE", "HEAD", "OPTIONS"},
		AllowCredentials: true,
	})

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", c.Handler(handler))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getBusPositions fetches bus positions from the MARTA API
func getBusPositions(apiURL string) []BusPosition {
	response, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch data from URL: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Failed to close response body: %v", err)
		}
	}(response.Body)

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	feed := &pb.FeedMessage{}
	err = proto.Unmarshal(data, feed)
	if err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}

	busPositions := make([]BusPosition, 0)

	for _, entity := range feed.Entity {
		position := entity.GetVehicle().GetPosition()
		currentStopSequence := entity.GetVehicle().GetCurrentStopSequence()
		stopId := entity.GetVehicle().GetStopId()
		currentStatus := entity.GetVehicle().GetCurrentStatus()
		timeStamp := entity.GetVehicle().GetTimestamp()
		congestionLevel := entity.GetVehicle().GetCongestionLevel()
		occupancyStatus := entity.GetVehicle().GetOccupancyStatus()

		vehiclePosition := VehiclePosition{
			Trip:                entity.GetVehicle().GetTrip(),
			Vehicle:             entity.GetVehicle().GetVehicle(),
			Position:            position,
			CurrentStopSequence: &currentStopSequence,
			StopId:              &stopId,
			CurrentStatus:       &currentStatus,
			Timestamp:           &timeStamp,
			CongestionLevel:     &congestionLevel,
			OccupancyStatus:     &occupancyStatus,
		}

		bus := BusPosition{
			ID:        vehiclePosition.Vehicle.GetId(),
			Latitude:  float64(vehiclePosition.Position.GetLatitude()),
			Longitude: float64(vehiclePosition.Position.GetLongitude()),
			Label:     vehiclePosition.Vehicle.GetLabel(),
			Bearing:   float64(vehiclePosition.Position.GetBearing()),
		}
		busPositions = append(busPositions, bus)
	}

	return busPositions
}

// getTripUpdates fetches trip updates from the MARTA API
func getTripUpdates(apiURL string) []TripUpdate {

	response, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch data from URL: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Failed to close response body: %v", err)
		}
	}(response.Body)

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	feed := &pb.FeedMessage{}
	err = proto.Unmarshal(data, feed)
	if err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}

	tripUpdates := make([]TripUpdate, 0)

	for _, entity := range feed.Entity {
		tripUpdate := entity.GetTripUpdate()
		stopTimeUpdate := tripUpdate.GetStopTimeUpdate()
		timestamp := tripUpdate.GetTimestamp()
		delay := tripUpdate.GetDelay()

		trip := TripUpdate{
			Trip:           tripUpdate.GetTrip(),
			Vehicle:        tripUpdate.GetVehicle(),
			StopTimeUpdate: stopTimeUpdate,
			Timestamp:      &timestamp,
			Delay:          &delay,
		}
		tripUpdates = append(tripUpdates, trip)
	}

	return tripUpdates

}

// ParseShapes parses a shapes.txt file and returns a slice of Shape structs.
func ParseShapes(filePath string) ([]Shape, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close file: %v", err)
		}
	}(file)

	shapes, err := ParseShapesFromReader(file)
	if err != nil {
		return nil, err
	}

	return shapes, nil
}

// ParseShapesFromReader parses a shapes.txt file and returns a slice of Shape structs.
func ParseShapesFromReader(file *os.File) ([]Shape, error) {

	newReader := csv.NewReader(file)
	newReader.FieldsPerRecord = -1

	records, err := newReader.ReadAll()
	if err != nil {
		return nil, err
	}

	shapes := make([]Shape, 0, len(records)-1)

	for _, record := range records[1:] {

		latitude, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}

		longitude, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, err
		}

		sequence, err := strconv.ParseInt(record[3], 10, 32)
		if err != nil {
			return nil, err
		}

		distTraveled, err := strconv.ParseFloat(record[4], 64)

		shape := Shape{
			ShapeId:      record[0],
			Latitude:     latitude,
			Longitude:    longitude,
			Sequence:     int(sequence),
			DistTraveled: distTraveled,
		}
		shapes = append(shapes, shape)
	}

	return shapes, nil
}

// ParseRoutes parses a routes.txt file and returns a slice of Route structs.
func ParseRoutes(filePath string) ([]Route, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close file: %v", err)
		}
	}(file)

	routes, err := ParseRoutesFromReader(file)
	if err != nil {
		return nil, err
	}

	return routes, nil
}

// ParseRoutesFromReader parses a routes.txt file and returns a slice of Route structs.
func ParseRoutesFromReader(file *os.File) ([]Route, error) {

	newReader := csv.NewReader(file)
	newReader.FieldsPerRecord = -1

	records, err := newReader.ReadAll()
	if err != nil {
		return nil, err

	}
	routes := make([]Route, 0, len(records))

	for _, record := range records {

		route := Route{
			ID:        record[0],
			ShortName: record[2],
			LongName:  record[3],
			Color:     record[7],
			TextColor: record[8],
		}
		routes = append(routes, route)
	}

	return routes, nil
}

func ParseStops(filePath string) ([]Stop, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close file: %v", err)
		}
	}(file)

	stops, err := ParseStopsFromReader(file)
	if err != nil {
		return nil, err
	}

	return stops, nil
}

func ParseStopsFromReader(file *os.File) ([]Stop, error) {
	newReader := csv.NewReader(file)
	newReader.FieldsPerRecord = -1

	records, err := newReader.ReadAll()
	if err != nil {
		return nil, err
	}

	stops := make([]Stop, 0, len(records)-1)

	for _, record := range records[1:] {
		latitude, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, err
		}
		longitude, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return nil, err
		}

		stop := Stop{
			StopID:    record[0],
			StopCode:  record[1],
			StopName:  record[2],
			StopDesc:  record[3],
			Latitude:  latitude,
			Longitude: longitude,
		}
		stops = append(stops, stop)
	}

	return stops, nil
}
