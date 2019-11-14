package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// SendSync sends a sync packet to the FSD server.
func (c *Context) SendSync() {
	err := fsd.Send(c.Consumer, "SYNC:*:"+c.Config.FSD.DataServerName+":B1:1:")
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
}
