package dataserver

import (
	"encoding/json"
	"fmt"
	"github.com/bugsnag/bugsnag-go"
	"github.com/olebedev/config"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	PilotData []PilotData `json:"pilots"`
	ATCData   []ATCData   `json:"controllers"`
}

// PilotData is data about individual pilots on the network.
type PilotData struct {
	Server     string     `json:"server"`
	Callsign   string     `json:"callsign"`
	Member     MemberData `json:"member"`
	Latitude   float64    `json:"latitude"`
	Longitude  float64    `json:"longitude"`
	Altitude   int32      `json:"altitude"`
	Speed      int32      `json:"speed"`
	Heading    int32      `json:"heading"`
	FlightPlan FlightPlan `json:"plan"`
}

// MemberData represents a user's personal data.
type MemberData struct {
	CID  int32  `json:"cid"`
	Name string `json:"name"`
}

// ATCData is data about individual controllers on the network.
type ATCData struct {
	Server       string     `json:"server"`
	Callsign     string     `json:"callsign"`
	Member       MemberData `json:"member"`
	Rating       int32      `json:"rating"`
	Frequency    string     `json:"frequency"`
	FacilityType int32      `json:"facility"`
	VisualRange  int32      `json:"range"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
}

// FlightPlan describes the data about a filed flight plan.
type FlightPlan struct {
	FlightRules string         `json:"flight_rules"`
	Aircraft    string         `json:"aircraft"`
	CruiseSpeed int32          `json:"cruise_speed"`
	Departure   string         `json:"departure"`
	Arrival     string         `json:"arrival"`
	Altitude    string         `json:"altitude"`
	Alternate   string         `json:"alternate"`
	Route       string         `json:"route"`
	Time        FlightPlanTime `json:"time"`
	Remarks     string         `json:"remarks"`
}

// FlightPlanTime represents the times present in a filed flight plan.
type FlightPlanTime struct {
	Departure string `json:"departure"`
	Enroute   string `json:"enroute"`
	Fuel      string `json:"fuel"`
}

// Cfg contains all of the necessary configuration data.
var Cfg *config.Config

// UpdatePosition updates a client's position data in the Client list and updates the JSON file.
func UpdatePosition(split []string, clientList *ClientList) error {
	fmt.Printf("Position Update Received: %v\n", split[6])
	latitude, err := convertStringToDouble(split[9])
	if err != nil {
		return errors.Wrapf(err, "Failed to get latitude %+v", split[9])
	}
	longitude, err := convertStringToDouble(split[10])
	if err != nil {
		return errors.Wrapf(err, "Failed to get longitude %+v", split[10])
	}
	altitude, err := convertStringToInteger(split[11])
	if err != nil {
		return errors.Wrapf(err, "Failed to get altitude %+v", split[11])
	}
	speed, err := convertStringToInteger(split[12])
	if err != nil {
		return errors.Wrapf(err, "Failed to get speed %+v", split[12])
	}
	heading, err := getHeading(split[13])
	if err != nil {
		return errors.Wrapf(err, "Failed to get heading %+v", split[13])
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == split[6] {
			*&clientList.PilotData[i].Latitude = latitude
			*&clientList.PilotData[i].Longitude = longitude
			*&clientList.PilotData[i].Altitude = altitude
			*&clientList.PilotData[i].Speed = speed
			*&clientList.PilotData[i].Heading = heading
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = writeDataFile(err, clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// UpdateFlightPlan updates the flight plan entry for the specified callsign
func UpdateFlightPlan(split []string, clientList *ClientList) error {
	fmt.Printf("Flight Plan Update Received: %v\n", split[5])
	cruiseSpeed, err := convertStringToInteger(split[9])
	if err != nil {
		return errors.Wrapf(err, "Failed to get cruise speed %+v", split[9])
	}
	departureTime, err := convertStringToInteger(split[11])
	if err != nil {
		return errors.Wrapf(err, "Failed to get departure time %+v", split[11])
	}
	enrouteTimeHours, err := convertStringToInteger(split[15])
	if err != nil {
		return errors.Wrapf(err, "Failed to get enroute time hours %+v", split[15])
	}
	enrouteTimeMinutes, err := convertStringToInteger(split[16])
	if err != nil {
		return errors.Wrapf(err, "Failed to get enroute time minutes %+v", split[16])
	}
	fuelTimeHours, err := convertStringToInteger(split[17])
	if err != nil {
		return errors.Wrapf(err, "Failed to get fuel time hours %+v", split[17])
	}
	fuelTimeMinutes, err := convertStringToInteger(split[18])
	if err != nil {
		return errors.Wrapf(err, "Failed to get fuel time minutes %+v", split[18])
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == split[5] {
			*&clientList.PilotData[i].FlightPlan = FlightPlan{
				FlightRules: split[7],
				Aircraft:    split[8],
				CruiseSpeed: cruiseSpeed,
				Departure:   split[10],
				Altitude:    split[13],
				Arrival:     split[14],
				Alternate:   split[19],
				Remarks:     split[20],
				Route:       split[21],
				Time: FlightPlanTime{
					Departure: fmt.Sprintf("%04d", departureTime),
					Enroute:   fmt.Sprintf("%02d", enrouteTimeHours) + fmt.Sprintf("%02d", enrouteTimeMinutes),
					Fuel:      fmt.Sprintf("%02d", fuelTimeHours) + fmt.Sprintf("%02d", fuelTimeMinutes),
				},
			}
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = writeDataFile(err, clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// convertStringToDouble converts a string to a float64
func convertStringToDouble(string string) (float64, error) {
	double, err := strconv.ParseFloat(string, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to convert string to float64 %+v", string)
	}
	return double, nil
}

// convertStringToInteger converts a string to an int32
func convertStringToInteger(string string) (int32, error) {
	integer, err := strconv.ParseInt(string, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to convert string to int32 %+v", string)
	}
	return int32(integer), nil
}

// getHeading parses the PBH FSD value to extract the heading
func getHeading(split string) (int32, error) {
	pbh, err := strconv.ParseUint(split, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse PBH field %+v", split)
	}
	hdgBit := (pbh >> 2) & 0x3FF
	heading := float64(hdgBit) / 1024.0 * 360
	if heading < 0.0 {
		heading += 360
	} else if heading >= 360.0 {
		heading -= 360
	}
	return int32(heading), nil
}

// UpdateControllerData updates a controllers's data in the Client list and updates the JSON file.
func UpdateControllerData(split []string, clientList *ClientList) error {
	fmt.Printf("Controller Update Received: %v\n", split[5])
	frequency, err := convertStringToDouble("1" + split[6][0:2] + "." + split[6][2:5])
	if err != nil {
		return errors.Wrapf(err, "Failed to get frequency %+v", split[6])
	}
	facilityType, err := convertStringToInteger(split[7])
	if err != nil {
		return errors.Wrapf(err, "Failed to get facility type %+v", split[7])
	}
	visualRange, err := convertStringToInteger(split[8])
	if err != nil {
		return errors.Wrapf(err, "Failed to get visual range %+v", split[8])
	}
	latitude, err := convertStringToDouble(split[10])
	if err != nil {
		return errors.Wrapf(err, "Failed to get latitude %+v", split[10])
	}
	longitude, err := convertStringToDouble(split[11])
	if err != nil {
		return errors.Wrapf(err, "Failed to get longitude %+v", split[11])
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == split[5] {
			*&clientList.ATCData[i].Frequency = fmt.Sprintf("%.3f", frequency)
			*&clientList.ATCData[i].FacilityType = facilityType
			*&clientList.ATCData[i].VisualRange = visualRange
			*&clientList.ATCData[i].Latitude = latitude
			*&clientList.ATCData[i].Longitude = longitude
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = writeDataFile(err, clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// RemoveClient removes a client from the Client list and updates the JSON file.
func RemoveClient(split []string, clientList *ClientList) error {
	fmt.Printf("Client Deleted: %v\n", split[5])
	for i, v := range clientList.PilotData {
		if v.Callsign == split[5] {
			*&clientList.PilotData = append(clientList.PilotData[:i], clientList.PilotData[i+1:]...)
			break
		}
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == split[5] {
			*&clientList.ATCData = append(clientList.ATCData[:i], clientList.ATCData[i+1:]...)
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = writeDataFile(err, clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// AddClient adds a client to the Client list and updates the JSON file.
func AddClient(split []string, clientList *ClientList) error {
	fmt.Printf("Client Added: %v\n", split[7])
	cid, err := convertStringToInteger(split[5])
	if err != nil {
		return errors.Wrapf(err, "Failed to get CID %+v", split[5])
	}
	rating, err := convertStringToInteger(split[9])
	if err != nil {
		return errors.Wrapf(err, "Failed to get rating %+v", split[9])
	}
	if split[8] == "1" {
		*&clientList.PilotData = append(clientList.PilotData, PilotData{
			Server:   split[6],
			Callsign: split[7],
			Member: MemberData{
				CID:  cid,
				Name: split[11],
			},
		})
	} else if split[8] == "2" {
		*&clientList.ATCData = append(clientList.ATCData, ATCData{
			Server:   split[6],
			Callsign: split[7],
			Rating:   rating,
			Member: MemberData{
				CID:  cid,
				Name: split[11],
			},
		})
	}
	clientJSON, err := encodeJSON(*clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = writeDataFile(err, clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// writeDataFile overwrites the data file with new data.
func writeDataFile(err error, clientJSON []byte) error {
	directory, err := Cfg.String("data.file.directory")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(directory+"vatsim-data.json", clientJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	return nil
}

// encodeJSON encodes the current Client list to JSON.
func encodeJSON(clientList ClientList) ([]byte, error) {
	clientJSON, err := json.Marshal(clientList)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	return clientJSON, nil
}

// ReadConfig reads the config file and instantiates the config object
func ReadConfig() {
	file, err := ioutil.ReadFile("configs/config.yml")
	if err != nil {
		panic(err)
	}
	yamlString := string(file)
	Cfg, err = config.ParseYaml(yamlString)
	if err != nil {
		panic(err)
	}
}

// ConfigureBugsnag sets up bugsnag for panic reporting
func ConfigureBugsnag() {
	apiKey, err := Cfg.String("bugsnag.credentials.api_key")
	if err != nil {
		panic(err)
	}
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: apiKey,
	})
}
