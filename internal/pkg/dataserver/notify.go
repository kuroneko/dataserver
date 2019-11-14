package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// SendNotify sends a notify packet to create our FSD server
func (c *Context) SendNotify() {
	notify := fsd.Notify{
		Base: fsd.Base{
			Destination:  "*",
			Source:       c.Config.FSD.DataServerName,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		FeedFlag: 0,
		Ident:    c.Config.FSD.DataServerName,
		Name:     c.Config.FSD.DataServerName,
		Email:    c.Config.FSD.DataServerEmail,
		Hostname: "127.0.0.1",
		Version:  "v1.0",
		Flags:    0,
		Location: c.Config.FSD.DataServerLocation,
	}
	err := fsd.Send(c.Consumer, notify.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
	log.WithField("packet", notify.Serialize()).Debug("Successfully sent NOTIFY packet.")
}
