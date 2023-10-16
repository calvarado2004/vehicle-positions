package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseShapes(t *testing.T) {
	// Create a temporary CSV file with sample data for testing
	tmpFile := createTempCSVFileForTesting("shapes.csv")
	defer func(tmpFile *os.File) {
		err := tmpFile.Close()
		if err != nil {
			t.Errorf("Error closing temporary CSV file: %v", err)
		}
	}(tmpFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Error removing temporary CSV file: %v", err)
		}
	}(tmpFile.Name())

	shapes, err := ParseShapes(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseShapes error: %v", err)
	}

	if len(shapes) != 2 {
		t.Errorf("Expected 2 shape, got %d", len(shapes))
	}

}

func createTempCSVFileForTesting(filename string) *os.File {
	// Create a temporary CSV file with sample data for testing
	tmpFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating temporary CSV file: %v", err)
	}

	writer := csv.NewWriter(tmpFile)
	err = writer.Write([]string{"shape_id", "latitude", "longitude", "sequence", "dist_traveled"})
	if err != nil {
		log.Printf("Error writing to temporary CSV file: %v\n", err)
		return nil
	}
	err = writer.Write([]string{"1", "37.123", "-122.456", "1", "0.0"})
	if err != nil {
		log.Printf("Error writing to temporary CSV file: %v\n", err)
		return nil
	}

	err = writer.Write([]string{"1", "37.456", "-122.789", "2", "2.0"})
	if err != nil {
		log.Printf("Error writing to temporary CSV file: %v\n", err)
		return nil
	}

	writer.Flush()

	return tmpFile
}

func TestGetBusPositions(t *testing.T) {
	// Read the static protobuf data from the file
	responseData, err := os.ReadFile("./test/vehiclepositions.pb")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err := w.Write(responseData)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	busPositions := getBusPositions(mockServer.URL)

	expectedID := "2301"
	if busPositions[0].ID != expectedID {
		t.Errorf("Expected ID %s, got %s", expectedID, busPositions[0].ID)
	}

	expectedLabel := "1601"
	if busPositions[0].Label != expectedLabel {
		t.Errorf("Expected label %s, got %s", expectedLabel, busPositions[0].Label)
	}

	expectedLength := 182
	if len(busPositions) != expectedLength {
		t.Errorf("Expected %d bus positions, got %d", expectedLength, len(busPositions))
	}

}

func TestParseRoutes(t *testing.T) {
	// Create a temporary CSV file with sample data for testing
	tmpFile := createTempCSVFileForRoutesTesting("routes.csv")
	defer func(tmpFile *os.File) {
		err := tmpFile.Close()
		if err != nil {
			t.Errorf("Error closing temporary CSV file: %v", err)
		}
	}(tmpFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Error removing temporary CSV file: %v", err)
		}
	}(tmpFile.Name())

	routes, err := ParseRoutes(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseRoutes error: %v", err)
	}

	// omitting the header row in the CSV file
	if len(routes)-1 != 4 {
		t.Errorf("Expected 4 routes, got %d", len(routes))
	}

	expectedRouteID := "20643"
	if routes[1].ID != expectedRouteID {
		t.Errorf("Expected route ID %s, got %s", expectedRouteID, routes[1].ID)
	}

	expectedRouteShortName := "1"
	if routes[1].ShortName != expectedRouteShortName {
		t.Errorf("Expected route short name %s, got %s", expectedRouteShortName, routes[1].ShortName)
	}

	expectedRouteLongName := "Marietta Blvd/Joseph E Lowery Blvd"
	if routes[1].LongName != expectedRouteLongName {
		t.Errorf("Expected route long name %s, got %s", expectedRouteLongName, routes[1].LongName)
	}

}

func createTempCSVFileForRoutesTesting(filename string) *os.File {
	// Create a temporary CSV file with sample data for testing
	tmpFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating temporary CSV file: %v", err)
	}

	writer := csv.NewWriter(tmpFile)
	err = writer.Write([]string{"route_id", "agency_id", "route_short_name", "route_long_name", "route_desc", "route_type", "route_url", "route_color", "route_text_color"})
	if err != nil {
		log.Printf("Error writing to temporary CSV file: %v\n", err)
		return nil
	}
	data := [][]string{
		{"20643", "MARTA", "1", "Marietta Blvd/Joseph E Lowery Blvd", "", "3", "https://itsmarta.com/1.aspx", "FF00FF", "000000"},
		{"20644", "MARTA", "2", "Ponce de Leon Avenue / Druid Hills", "", "3", "https://itsmarta.com/2.aspx", "008000", "000000"},
		{"20645", "MARTA", "3", "Martin Luther King Jr Dr/Auburn Ave", "", "3", "https://itsmarta.com/3.aspx", "FF8000", "000000"},
		{"20646", "MARTA", "4", "Moreland Avenue", "", "3", "https://itsmarta.com/4.aspx", "FF00FF", "000000"},
	}
	for _, record := range data {
		err := writer.Write(record)
		if err != nil {
			log.Printf("Error writing to temporary CSV file: %v\n", err)
			return nil
		}
	}

	writer.Flush()

	return tmpFile
}

func TestParseStops(t *testing.T) {
	tmpFile := createTempCSVFileForStopsTesting("stops.csv")
	defer func(tmpFile *os.File) {
		err := tmpFile.Close()
		if err != nil {
			t.Errorf("Error closing temporary CSV file: %v", err)
		}
	}(tmpFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Errorf("Error removing temporary CSV file: %v", err)
		}
	}(tmpFile.Name())

	stops, err := ParseStops(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseStops error: %v", err)
	}

	if len(stops) != 3 {
		t.Errorf("Expected 3 stops, got %d", len(stops))
	}

	expectedStopID := "27"
	if stops[0].StopID != expectedStopID {
		t.Errorf("Expected stop ID %s, got %s", expectedStopID, stops[0].StopID)
	}

	expectedStopCode := "907933"
	if stops[0].StopCode != expectedStopCode {
		t.Errorf("Expected stop code %s, got %s", expectedStopCode, stops[0].StopCode)
	}

	expectedStopName := "HAMILTON E HOLMES STATION"
	if stops[0].StopName != expectedStopName {
		t.Errorf("Expected stop name %s, got %s", expectedStopName, stops[0].StopName)
	}

	expectedStopDesc := "70 HAMILTON E HOLMES DR NW & CSX TRANSPORTATION"
	if stops[0].StopDesc != expectedStopDesc {
		t.Errorf("Expected stop description %s, got %s", expectedStopDesc, stops[0].StopDesc)
	}

}

func createTempCSVFileForStopsTesting(filename string) *os.File {
	tmpFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating temporary CSV file: %v", err)
	}

	writer := csv.NewWriter(tmpFile)
	err = writer.Write([]string{"stop_id", "stop_code", "stop_name", "stop_desc", "stop_lat", "stop_lon", "zone_id", "stop_url", "location_type", "parent_station", "stop_timezone", "wheelchair_boarding"})
	if err != nil {
		log.Fatalf("Error writing to temporary CSV file: %v", err)
	}
	data := [][]string{
		{"27", "907933", "HAMILTON E HOLMES STATION", "70 HAMILTON E HOLMES DR NW & CSX TRANSPORTATION", "33.754553", "-84.469302", "", "", "", "", "", "1"},
		{"28", "908023", "WEST LAKE STATION", "80 ANDERSON AVE NW & CSX TRANSPORTATION", "33.753328", "-84.445329", "", "", "", "", "", "1"},
		{"39", "907906", "WEST LAKE STATION", "80 ANDERSON AVE NW & CSX TRANSPORTATION", "33.753247", "-84.445568", "", "", "", "", "", "1"},
	}
	for _, record := range data {
		err := writer.Write(record)
		if err != nil {
			log.Fatalf("Error writing to temporary CSV file: %v", err)
		}
	}

	writer.Flush()

	return tmpFile
}

func TestGetTripUpdates(t *testing.T) {
	responseData, err := os.ReadFile("./test/tripupdates.pb")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err := w.Write(responseData)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	tripUpdates := getTripUpdates(mockServer.URL)

	expectedTripID := "8729521"
	if tripUpdates[0].Trip.GetTripId() != expectedTripID {
		t.Errorf("Expected trip ID %s, got %s", expectedTripID, tripUpdates[0].Trip)
	}

	expectedRouteID := "20708"
	if tripUpdates[0].Trip.GetRouteId() != expectedRouteID {
		t.Errorf("Expected route ID %s, got %s", expectedRouteID, tripUpdates[0].Trip)
	}

	expectedStopID := "42100"
	if tripUpdates[0].StopTimeUpdate[0].GetStopId() != expectedStopID {
		t.Errorf("Expected stop ID %s, got %s", expectedStopID, tripUpdates[0].StopTimeUpdate[0])
	}

	expectedArrivalDelay := int32(0)
	if tripUpdates[0].StopTimeUpdate[0].GetArrival().GetDelay() != expectedArrivalDelay {
		t.Errorf("Expected arrival delay %d, got %d", expectedArrivalDelay, tripUpdates[0].StopTimeUpdate[0].GetArrival().GetDelay())
	}

}
