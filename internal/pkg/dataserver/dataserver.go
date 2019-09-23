package dataserver

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/olebedev/config"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
	"time"
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
	Latitude   float32    `json:"latitude"`
	Longitude  float32    `json:"longitude"`
	Altitude   int        `json:"altitude"`
	Speed      int        `json:"speed"`
	Heading    int        `json:"heading"`
	FlightPlan FlightPlan `json:"plan"`
}

// MemberData represents a user's personal data.
type MemberData struct {
	CID  int    `json:"cid"`
	Name string `json:"name"`
}

// ATCData is data about individual controllers on the network.
type ATCData struct {
	Server       string     `json:"server"`
	Callsign     string     `json:"callsign"`
	Member       MemberData `json:"member"`
	Rating       int        `json:"rating"`
	Frequency    string     `json:"frequency"`
	FacilityType int        `json:"facility"`
	VisualRange  int        `json:"range"`
	Latitude     float32    `json:"latitude"`
	Longitude    float32    `json:"longitude"`
}

// FlightPlan describes the data about a filed flight plan.
type FlightPlan struct {
	FlightRules string         `json:"flight_rules"`
	Aircraft    string         `json:"aircraft"`
	CruiseSpeed int            `json:"cruise_speed"`
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

var (
	// Cfg contains all of the necessary configuration data.
	Cfg *config.Config

	// Channel streams the clientList updates
	Channel = make(chan ClientList)
)

// UpdatePosition updates a client's position data in the Client list and updates the JSON file.
func UpdatePosition(split []string, clientList *ClientList) error {
	fmt.Printf("%+v Position Update Received: %+v\n", time.Now().UTC().Format(time.RFC3339), split[6])
	latitude, err := convertStringToDouble(split[9])
	if err != nil {
		return errors.Wrapf(err, "Failed to get latitude %+v", split[9])
	}
	longitude, err := convertStringToDouble(split[10])
	if err != nil {
		return errors.Wrapf(err, "Failed to get longitude %+v", split[10])
	}
	altitude, err := strconv.Atoi(split[11])
	if err != nil {
		return errors.Wrapf(err, "Failed to get altitude %+v", split[11])
	}
	speed, err := strconv.Atoi(split[12])
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
	Channel <- *clientList
	return nil
}

// UpdateFlightPlan updates the flight plan entry for the specified callsign
func UpdateFlightPlan(split []string, clientList *ClientList) error {
	fmt.Printf("%+v Flight Plan Update Received: %+v\n", time.Now().UTC().Format(time.RFC3339), split[5])
	cruiseSpeed, err := strconv.Atoi(split[9])
	if err != nil {
		return errors.Wrapf(err, "Failed to get cruise speed %+v", split[9])
	}
	departureTime, err := strconv.Atoi(split[11])
	if err != nil {
		return errors.Wrapf(err, "Failed to get departure time %+v", split[11])
	}
	enrouteTimeHours, err := strconv.Atoi(split[15])
	if err != nil {
		return errors.Wrapf(err, "Failed to get enroute time hours %+v", split[15])
	}
	enrouteTimeMinutes, err := strconv.Atoi(split[16])
	if err != nil {
		return errors.Wrapf(err, "Failed to get enroute time minutes %+v", split[16])
	}
	fuelTimeHours, err := strconv.Atoi(split[17])
	if err != nil {
		return errors.Wrapf(err, "Failed to get fuel time hours %+v", split[17])
	}
	fuelTimeMinutes, err := strconv.Atoi(split[18])
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
	Channel <- *clientList
	return nil
}

// convertStringToDouble converts a string to a float64
func convertStringToDouble(string string) (float32, error) {
	double, err := strconv.ParseFloat(string, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to convert string to float32 %+v", string)
	}
	return float32(double), nil
}

// getHeading parses the PBH FSD value to extract the heading
func getHeading(split string) (int, error) {
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
	return int(heading), nil
}

// UpdateControllerData updates a controllers's data in the Client list and updates the JSON file.
func UpdateControllerData(split []string, clientList *ClientList) error {
	fmt.Printf("%+v Controller Update Received: %+v\n", time.Now().UTC().Format(time.RFC3339), split[5])
	var frequency float32
	var err error
	if len(split[6]) >= 5 {
		frequency, err = convertStringToDouble("1" + split[6][0:2] + "." + split[6][2:5])
		if err != nil {
			return errors.Wrapf(err, "Failed to get frequency %+v", split[6])
		}
	} else {
		return errors.New("Invalid frequency: " + split[6])
	}
	facilityType, err := strconv.Atoi(split[7])
	if err != nil {
		return errors.Wrapf(err, "Failed to get facility type %+v", split[7])
	}
	visualRange, err := strconv.Atoi(split[8])
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
	Channel <- *clientList
	return nil
}

// RemoveClient removes a client from the Client list and updates the JSON file.
func RemoveClient(split []string, clientList *ClientList) error {
	fmt.Printf("%+v Client Deleted: %+v\n", time.Now().UTC().Format(time.RFC3339), split[5])
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
	Channel <- *clientList
	return nil
}

// AddClient adds a client to the Client list and updates the JSON file.
func AddClient(split []string, clientList *ClientList) error {
	fmt.Printf("%+v Client Added: %+v\n", time.Now().UTC().Format(time.RFC3339), split[7])
	cid, err := strconv.Atoi(split[5])
	if err != nil {
		return errors.Wrapf(err, "Failed to get CID %+v", split[5])
	}
	rating, err := strconv.Atoi(split[9])
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
	Channel <- *clientList
	return nil
}

// WriteDataFile overwrites the data file with new data.
func WriteDataFile(clientJSON []byte) error {
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

// EncodeJSON encodes the current Client list to JSON.
func EncodeJSON(clientList ClientList) ([]byte, error) {
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

// ConfigureSentry sets up bugsnag for panic reporting
func ConfigureSentry() {
	dsn, err := Cfg.String("sentry.credentials.dsn")
	if err != nil {
		panic(err)
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		panic(err)
	}
}
