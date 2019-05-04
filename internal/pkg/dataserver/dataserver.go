package dataserver

import (
	"encoding/json"
	"fmt"
	"github.com/olebedev/config"
	"io/ioutil"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	ClientData []ClientData `json:"data"`
}

// ClientData is data about individual client's on the network.
type ClientData struct {
	Server   string       `json:"server"`
	CID      string       `json:"cid"`
	Callsign string       `json:"callsign"`
	Rating   string       `json:"rating"`
	Name     string       `json:"name"`
	Position PositionData `json:"position"`
}

// PositionData describes data about an individual client's geographic position.
type PositionData struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Altimeter string `json:"altimeter"`
	Speed     string `json:"speed"`
}

// cfg contains all of the necessary configuration data.
var Cfg *config.Config

// UpdatePosition updates a client's position data in the Client list and updates the JSON file.
func UpdatePosition(split []string, clientList *ClientList) {
	fmt.Printf("Position Update Received: %v\n", split[6])
	for i, v := range clientList.ClientData {
		if v.Callsign == split[6] {
			*&clientList.ClientData[i].Position = PositionData{
				Latitude:  split[9],
				Longitude: split[10],
				Altimeter: split[11],
				Speed:     split[12],
			}
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	writeDataFile(err, clientJSON)
}

// RemoveClient removes a client from the Client list and updates the JSON file.
func RemoveClient(split []string, clientList *ClientList) {
	fmt.Printf("Client Deleted: %v\n", split[5])
	for i, v := range clientList.ClientData {
		if v.Callsign == split[5] {
			*&clientList.ClientData = append(clientList.ClientData[:i], clientList.ClientData[i+1:]...)
			break
		}
	}
	clientJSON, err := encodeJSON(*clientList)
	writeDataFile(err, clientJSON)
}

// AddClient adds a client to the Client list and updates the JSON file.
func AddClient(split []string, clientList *ClientList) {
	fmt.Printf("Client Added: %v\n", split[7])
	*&clientList.ClientData = append(clientList.ClientData, ClientData{
		Server:   split[6],
		CID:      split[5],
		Callsign: split[7],
		Rating:   split[9],
		Name:     split[11],
	})
	clientJSON, err := encodeJSON(*clientList)
	writeDataFile(err, clientJSON)
}

// writeDataFile overwrites the data file with new data.
func writeDataFile(err error, clientJSON []byte) {
	err = ioutil.WriteFile("vatsim-data.json", clientJSON, 0644)
	if err != nil {
		panic(err)
	}
}

// encodeJSON encodes the current Client list to JSON.
func encodeJSON(clientList ClientList) ([]byte, error) {
	clientJSON, err := json.Marshal(clientList)
	if err != nil {
		panic(err)
	}
	return clientJSON, err
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
