package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// SendSync sends a sync packet to the FSD server.
func (c *Context) SendSync() {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.Fatal("Data server name not defined.")
	}
	err = fsd.Send(c.Consumer, "SYNC:*:"+name+":B1:1:")
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
}
