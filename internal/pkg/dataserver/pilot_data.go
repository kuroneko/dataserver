package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// PilotData is data about individual pilots on the network.
type PilotData struct {
	Server     string     `json:"server"`
	Callsign   string     `json:"callsign"`
	Member     MemberData `json:"member"`
	Latitude   float64    `json:"latitude"`
	Longitude  float64    `json:"longitude"`
	Altitude   int        `json:"altitude"`
	Speed      int        `json:"speed"`
	Heading    int        `json:"heading"`
	FlightPlan FlightPlan `json:"plan"`
}

// HandlePilotData updates a client's position data in the Client list and updates the JSON file.
func HandlePilotData(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	pilotData, err := fsd.DeserializePilotData(fields)
	if err != nil {
		return err
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == fields[6] {
			*&clientList.PilotData[i].Latitude = pilotData.Latitude
			*&clientList.PilotData[i].Longitude = pilotData.Longitude
			*&clientList.PilotData[i].Altitude = pilotData.Altitude
			*&clientList.PilotData[i].Speed = pilotData.GroundSpeed
			*&clientList.PilotData[i].Heading = pilotData.Heading
			kafkaPush(producer, clientList.PilotData[i], "update_position")
			break
		}
	}
	log.WithFields(log.Fields{
		"callsign":  pilotData.Callsign,
		"latitude":  pilotData.Latitude,
		"longitude": pilotData.Longitude,
		"altitude":  pilotData.Altitude,
		"speed":     pilotData.GroundSpeed,
		"heading":   pilotData.Heading,
	}).Info("Pilot data packet received.")
	Channel <- *clientList
	return nil
}
