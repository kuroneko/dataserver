package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// HandleATISData updates controller information
func HandleATISData(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	atis, err := fsd.DeserializeATISData(fields)
	if err != nil {
		return err
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == atis.From {
			switch atis.Type {
			case "T":
				handleATISText(v, clientList, i, atis)
				break
			case "E":
				*&clientList.ATCData[i].ATISReceived = true
			}
			kafkaPush(producer, clientList.ATCData[i], "update_controller_data")
		}
	}
	log.WithFields(log.Fields{
		"callsign": atis.From,
		"data":     atis.Data,
	}).Info("ATIS data packet received.")
	Channel <- *clientList
	return nil
}

// handleATISText updates a controller's information
func handleATISText(v ATCData, clientList *ClientList, i int, atis fsd.ATISData) {
	switch v.ATISReceived {
	case true:
		*&clientList.ATCData[i].ATIS = atis.Data
		*&clientList.ATCData[i].ATISReceived = false
		break
	case false:
		*&clientList.ATCData[i].ATIS += "\n" + atis.Data
		break
	}
}
