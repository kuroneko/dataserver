package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net/textproto"
)

// ATCData is data about individual controllers on the network.
type ATCData struct {
	Server       string     `json:"server"`
	Callsign     string     `json:"callsign"`
	Member       MemberData `json:"member"`
	Rating       int        `json:"rating"`
	Frequency    int        `json:"frequency"`
	FacilityType int        `json:"facility"`
	VisualRange  int        `json:"range"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	ATIS         string     `json:"atis"`
	ATISReceived bool       `json:"-"`
}

// HandleATCData updates a controllers's data in the Client list and updates the JSON file.
func HandleATCData(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	atcData, err := fsd.DeserializeATCData(fields)
	if err != nil {
		return err
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == atcData.Callsign {
			*&clientList.ATCData[i].Frequency = atcData.Frequency
			*&clientList.ATCData[i].FacilityType = atcData.FacilityType
			*&clientList.ATCData[i].VisualRange = atcData.VisualRange
			*&clientList.ATCData[i].Latitude = atcData.Latitude
			*&clientList.ATCData[i].Longitude = atcData.Longitude
			kafkaPush(producer, clientList.ATCData[i], "update_controller_data")
			break
		}
	}
	log.WithFields(log.Fields{
		"callsign":  atcData.Callsign,
		"latitude":  atcData.Latitude,
		"longitude": atcData.Longitude,
	}).Info("ATC data packet received.")
	Channel <- *clientList
	return nil
}

// sendATCData sends fake ATC data for our client
func sendATCData(name string, err error, conn *textproto.Conn) {
	atcData := fsd.ATCData{
		Base: fsd.Base{
			Destination:  "*",
			Source:       name,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		Callsign:     name,
		Frequency:    99999,
		FacilityType: 1,
		VisualRange:  100,
		Rating:       1,
		Latitude:     0.00000,
		Longitude:    0.00000,
	}
	err = fsd.Send(conn, atcData.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": conn,
			"error":      err,
		}).Panic("Failed to send AD packet to FSD server.")
	}
	log.WithField("packet", atcData.Serialize()).Info("Successfully sent AD packet to server.")
}
