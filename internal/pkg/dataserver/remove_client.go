package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// RemoveClient removes a client from the Client list and updates the JSON file.
func (c *Context) RemoveClient(fields []string) error {
	removeClient, err := fsd.DeserializeRemoveClient(fields)
	if err != nil {
		return err
	}
	c.ClientList.Mutex.Lock()
	defer c.ClientList.Mutex.Unlock()
	for i, v := range c.ClientList.PilotData {
		if v.Callsign == removeClient.Callsign {
			*&c.ClientList.PilotData = append(c.ClientList.PilotData[:i], c.ClientList.PilotData[i+1:]...)
			data := Pilot{
				Server:   "",
				Callsign: removeClient.Callsign,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}

			kafkaPush(c.Producer, data, "remove_client")
			break
		}
	}
	for i, v := range c.ClientList.ATCData {
		if v.Callsign == removeClient.Callsign {
			*&c.ClientList.ATCData = append(c.ClientList.ATCData[:i], c.ClientList.ATCData[i+1:]...)
			data := ATC{
				Server:   "",
				Callsign: removeClient.Callsign,
				Rating:   0,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}
			kafkaPush(c.Producer, data, "remove_client")
			break
		}
	}
	totalConnections.With(prometheus.Labels{"server": removeClient.Source}).Dec()
	log.WithFields(log.Fields{
		"callsign": removeClient.Callsign,
		"server":   removeClient.Source,
	}).Info("Remove client packet received.")
	Channel <- *c.ClientList
	return nil
}
