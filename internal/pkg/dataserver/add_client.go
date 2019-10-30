package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net/textproto"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	PilotData []PilotData `json:"pilots"`
	ATCData   []ATCData   `json:"controllers"`
}

// MemberData represents a user's personal data.
type MemberData struct {
	CID  int    `json:"cid"`
	Name string `json:"name"`
}

// HandleAddClient adds a client to the Client list and updates the JSON file.
func HandleAddClient(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	addClient, err := fsd.DeserializeAddClient(fields)
	if err != nil {
		return err
	}
	if addClient.Type == 1 {
		data := PilotData{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&clientList.PilotData = append(clientList.PilotData, data)
		kafkaPush(producer, data, "add_client")
	} else if addClient.Type == 2 {
		data := ATCData{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Rating:   addClient.Rating,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&clientList.ATCData = append(clientList.ATCData, data)
		kafkaPush(producer, data, "add_client")
	}
	totalConnections.With(prometheus.Labels{"server": addClient.Server}).Inc()
	log.WithFields(log.Fields{
		"callsign": addClient.Callsign,
		"name":     addClient.RealName,
		"server":   addClient.Source,
	}).Info("Add client packet received.")
	Channel <- *clientList
	return nil
}

// sendAddClient sends the packet to connect our fake client
func sendAddClient(name string, err error, conn *textproto.Conn) {
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
	err = fsd.Send(conn, addClient.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": conn,
			"error":      err,
		}).Panic("Failed to send ADDCLIENT packet to FSD server.")
	}
	log.WithField("packet", addClient.Serialize()).Info("Successfully sent ADDCLIENT packet to server.")
}
