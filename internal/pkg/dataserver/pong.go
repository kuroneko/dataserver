package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// sendPong responds to a ping echoing back the data
func (c *Context) sendPong(ping fsd.Ping) {
	pong := fsd.Pong{
		Base: fsd.Base{
			Destination:  ping.Source,
			Source:       c.Config.FSD.DataServerName,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		Data: ping.Data,
	}
	err := fsd.Send(c.Consumer, pong.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send PONG packet to FSD server.")
	}
	log.WithField("packet", pong.Serialize()).Debug("Successfully sent PONG packet to server.")
}
