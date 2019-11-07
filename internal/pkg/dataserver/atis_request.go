package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// sendATISRequest sends the ATIS request packet to all ATC clients
func (c *Context) sendATISRequest(name string) {
	for _, v := range c.ClientList.ATCData {
		atisRequest := fsd.ATISRequest{
			Base: fsd.Base{
				Destination:  v.Callsign,
				Source:       name,
				PacketNumber: fsd.PdCount,
				HopCount:     1,
			},
			From: name,
		}
		err := fsd.Send(c.Consumer, atisRequest.Serialize())
		if err != nil {
			log.WithField("packet", atisRequest.Serialize()).Error("Failed to request ATIS.")
		}
		log.WithField("packet", atisRequest.Serialize()).Debug("Successfully requested ATIS.")
	}
}
