package dataserver

import (
	"encoding/json"
	"fmt"
	"github.com/bugsnag/bugsnag-go"
	"github.com/olebedev/config"
	"github.com/pkg/errors"
	"io/ioutil"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	PilotData []PilotData `json:"pilot"`
	ATCData   []ATCData   `json:"atc"`
}

// PilotData is data about individual pilots on the network.
type PilotData struct {
	Server   string       `json:"server"`
	CID      string       `json:"cid"`
	Callsign string       `json:"callsign"`
	Rating   string       `json:"rating"`
	Name     string       `json:"name"`
	Position PositionData `json:"position"`
}

// ATCData is data about individual controllers on the network.
type ATCData struct {
	Server     string         `json:"server"`
	CID        string         `json:"cid"`
	Callsign   string         `json:"callsign"`
	Rating     string         `json:"rating"`
	Name       string         `json:"name"`
	Controller ControllerData `json:"controller"`
}

// ControllerData describes data about an individual controller's setup.
type ControllerData struct {
	Frequency    string `json:"frequency"`
	FacilityType string `json:"facility"`
	VisualRange  string `json:"range"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
}

// PositionData describes data about an individual pilots's geographic position.
type PositionData struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Altimeter string `json:"altimeter"`
	Speed     string `json:"speed"`
}

// Cfg contains all of the necessary configuration data.
var Cfg *config.Config

// UpdatePosition updates a client's position data in the Client list and updates the JSON file.
func UpdatePosition(split []string, clientList *ClientList) error {
	fmt.Printf("Position Update Received: %v\n", split[6])
	for i, v := range clientList.PilotData {
		if v.Callsign == split[6] {
			*&clientList.PilotData[i].Position = PositionData{
				Latitude:  split[9],
				Longitude: split[10],
				Altimeter: split[11],
				Speed:     split[12],
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

// UpdateControllerData updates a controllers's data in the Client list and updates the JSON file.
func UpdateControllerData(split []string, clientList *ClientList) error {
	fmt.Printf("Controller Update Received: %v\n", split[5])
	for i, v := range clientList.ATCData {
		if v.Callsign == split[5] {
			*&clientList.ATCData[i].Controller = ControllerData{
				Frequency:    "1" + split[6][0:2] + "." + split[6][2:3],
				FacilityType: split[7],
				VisualRange:  split[8],
				Latitude:     split[10],
				Longitude:    split[11],
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
	if split[8] == "1" {
		*&clientList.PilotData = append(clientList.PilotData, PilotData{
			Server:   split[6],
			CID:      split[5],
			Callsign: split[7],
			Rating:   split[9],
			Name:     split[11],
		})
	} else if split[8] == "2" {
		*&clientList.ATCData = append(clientList.ATCData, ATCData{
			Server:   split[6],
			CID:      split[5],
			Callsign: split[7],
			Rating:   split[9],
			Name:     split[11],
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
