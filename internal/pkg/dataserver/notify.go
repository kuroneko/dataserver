package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// SendNotify sends a notify packet to create our FSD server
func (c *Context) SendNotify() {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.Fatal("Data server name not defined.")
	}
	email, err := config.Cfg.String("data.server.email")
	if err != nil {
		log.Fatal("Data server email not defined.")
	}
	location, err := config.Cfg.String("data.server.location")
	if err != nil {
		log.Fatal("Data server location not defined.")
	}
	notify := fsd.Notify{
		Base: fsd.Base{
			Destination:  "*",
			Source:       name,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		FeedFlag: 0,
		Ident:    name,
		Name:     name,
		Email:    email,
		Hostname: "127.0.0.1",
		Version:  "v1.0",
		Flags:    0,
		Location: location,
	}
	err = fsd.Send(c.Consumer, notify.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
	log.WithField("packet", notify.Serialize()).Debug("Successfully sent NOTIFY packet.")
}
