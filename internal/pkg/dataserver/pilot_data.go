package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"time"
)

// Pilot is data about individual pilots on the network.
type Pilot struct {
	Server      string     `json:"server"`
	Callsign    string     `json:"callsign"`
	Member      MemberData `json:"member"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Altitude    int        `json:"altitude"`
	Speed       int        `json:"speed"`
	Heading     int        `json:"heading"`
	FlightPlan  FlightPlan `json:"plan"`
	LastUpdated time.Time  `json:"last_updated"`
}

// HandlePilotData updates a client's position data in the Client list and updates the JSON file.
func (c *Context) HandlePilotData(fields []string) error {
	pilotData, err := fsd.DeserializePilotData(fields)
	if err != nil {
		return err
	}
	for i, v := range c.ClientList.PilotData {
		if v.Callsign == fields[6] {
			*&c.ClientList.PilotData[i].Latitude = pilotData.Latitude
			*&c.ClientList.PilotData[i].Longitude = pilotData.Longitude
			*&c.ClientList.PilotData[i].Altitude = pilotData.Altitude
			*&c.ClientList.PilotData[i].Speed = pilotData.GroundSpeed
			*&c.ClientList.PilotData[i].Heading = pilotData.Heading
			*&c.ClientList.PilotData[i].LastUpdated = time.Now().UTC()
			kafkaPush(c.Producer, c.ClientList.PilotData[i], "update_position")
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
	}).Debug("Pilot data packet received.")
	Channel <- *c.ClientList
	return nil
}
