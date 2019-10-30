package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// RemoveClient removes a client from the Client list and updates the JSON file.
func RemoveClient(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	removeClient, err := fsd.DeserializeRemoveClient(fields)
	if err != nil {
		return err
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == removeClient.Callsign {
			*&clientList.PilotData = append(clientList.PilotData[:i], clientList.PilotData[i+1:]...)
			data := PilotData{
				Server:   "",
				Callsign: removeClient.Callsign,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}

			kafkaPush(producer, data, "remove_client")
			break
		}
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == removeClient.Callsign {
			*&clientList.ATCData = append(clientList.ATCData[:i], clientList.ATCData[i+1:]...)
			data := ATCData{
				Server:   "",
				Callsign: removeClient.Callsign,
				Rating:   0,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}
			kafkaPush(producer, data, "remove_client")
			break
		}
	}
	totalConnections.With(prometheus.Labels{"server": removeClient.Source}).Dec()
	log.WithFields(log.Fields{
		"callsign": removeClient.Callsign,
		"server":   removeClient.Source,
	}).Info("Remove client packet received.")
	Channel <- *clientList
	return nil
}
