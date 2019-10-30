package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"net/textproto"
)

// sendATISRequest sends the ATIS request packet to all ATC clients
func sendATISRequest(clientList *ClientList, name string, conn *textproto.Conn) {
	for _, v := range clientList.ATCData {
		atisRequest := fsd.ATISRequest{
			Base: fsd.Base{
				Destination:  v.Callsign,
				Source:       name,
				PacketNumber: fsd.PdCount,
				HopCount:     1,
			},
			From: name,
		}
		err := fsd.Send(conn, atisRequest.Serialize())
		if err != nil {
			log.WithField("packet", atisRequest.Serialize()).Error("Failed to request ATIS.")
			sentry.CaptureException(err)
		}
		log.WithField("packet", atisRequest.Serialize()).Info("Successfully requested ATIS.")
	}
}
