package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"net/textproto"
)

// sendPong responds to a ping echoing back the data
func sendPong(ping fsd.Ping, conn *textproto.Conn) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Panic("Failed to retrieve server name from config file.")
	}
	pong := fsd.Pong{
		Base: fsd.Base{
			Destination:  ping.Source,
			Source:       name,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		Data: ping.Data,
	}
	err = fsd.Send(conn, pong.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": conn,
			"error":      err,
		}).Panic("Failed to send PONG packet to FSD server.")
	}
	log.WithField("packet", pong.Serialize()).Info("Successfully sent PONG packet to server.")
}
