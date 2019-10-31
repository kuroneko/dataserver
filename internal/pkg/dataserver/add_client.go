package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sync"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	PilotData []Pilot       `json:"pilots"`
	ATCData   []ATC         `json:"controllers"`
	Mutex     *sync.RWMutex `json:"-"`
}

// MemberData represents a user's personal data.
type MemberData struct {
	CID  int    `json:"cid"`
	Name string `json:"name"`
}

// HandleAddClient adds a client to the Client list and updates the JSON file.
func (c *Context) HandleAddClient(fields []string) error {
	addClient, err := fsd.DeserializeAddClient(fields)
	if err != nil {
		return err
	}
	c.ClientList.Mutex.Lock()
	defer c.ClientList.Mutex.Unlock()
	if addClient.Type == 1 {
		data := Pilot{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&c.ClientList.PilotData = append(c.ClientList.PilotData, data)
		kafkaPush(c.Producer, data, "add_client")
	} else if addClient.Type == 2 {
		data := ATC{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Rating:   addClient.Rating,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&c.ClientList.ATCData = append(c.ClientList.ATCData, data)
		kafkaPush(c.Producer, data, "add_client")
	}
	totalConnections.With(prometheus.Labels{"server": addClient.Server}).Inc()
	log.WithFields(log.Fields{
		"callsign": addClient.Callsign,
		"name":     addClient.RealName,
		"server":   addClient.Source,
	}).Debug("Add client packet received.")
	Channel <- *c.ClientList
	return nil
}

// sendAddClient sends the packet to connect our fake client
func (c *Context) sendAddClient(name string) {
	addClient := fsd.AddClient{
		Base: fsd.Base{
			Destination:  "*",
			Source:       name,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		CID:              0,
		Server:           name,
		Callsign:         name,
		Type:             2,
		Rating:           1,
		ProtocolRevision: 100,
		RealName:         name,
		SimType:          -1,
		Hidden:           1,
	}
	err := fsd.Send(c.Consumer, addClient.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send ADDCLIENT packet to FSD server.")
	}
	log.WithField("packet", addClient.Serialize()).Debug("Successfully sent ADDCLIENT packet to server.")
}
