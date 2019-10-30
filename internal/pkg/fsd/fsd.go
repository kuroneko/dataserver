package fsd

import (
	"dataserver/internal/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/textproto"
	"strings"
)

// pdCount is a count of the number of packets we have sent to the FSD server.
var PdCount int

// Connect establishes a connection to the FSD server.
func Connect() *textproto.Conn {
	ip, err := config.Cfg.String("fsd.server.ip")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get FSD IP from configuration.")
	}
	port, err := config.Cfg.String("fsd.server.port")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get FSD port from configuration.")
	}
	conn, err := textproto.Dial("tcp", ip+":"+port)
	if err != nil {
		log.WithFields(log.Fields{
			"ip":    ip,
			"port":  port,
			"error": err,
		}).Fatal("Failed to connect to FSD server.")
	}
	return conn
}

// send formats and sends a new FSD packet to the FSD server.
func Send(conn *textproto.Conn, message string) error {
	_, err := conn.Cmd(message)
	if err != nil {
		return errors.Wrapf(err, "Failed to send packet to FSD server. %+v", conn)
	}
	PdCount++
	return nil
}

// ParseMessage splits an FSD message based on the colon delimiter for further handling.
func ParseMessage(message string) []string {
	split := strings.Split(message, ":")
	return split
}

// ReadMessage reads a new FSD message from the connection.
func ReadMessage(conn *textproto.Conn) (string, error) {
	message, err := conn.ReadLine()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read new FSD message from connection. %+v", conn)
	}
	return message, nil
}

// Sync sends a sync packet to the FSD server.
func Sync(conn *textproto.Conn) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get server name from configuration.")
	}
	err = Send(conn, "SYNC:*:"+name+":B1:1:")
	if err != nil {
		log.WithFields(log.Fields{
			"connection": conn,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
}

// SendNotify sends a notify packet to create our FSD server
func SendNotify(conn *textproto.Conn) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get dataserver name from configuration.")
	}
	email, err := config.Cfg.String("data.server.email")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get dataserver email from configuration.")
	}
	location, err := config.Cfg.String("data.server.location")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Fatal("Failed to get dataserver location from configuration.")
	}
	notify := Notify{
		Base: Base{
			Destination:  "*",
			Source:       name,
			PacketNumber: PdCount,
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
	err = Send(conn, notify.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": conn,
			"error":      err,
		}).Fatal("Failed to send SYNC packet to FSD server.")
	}
	log.WithField("packet", notify.Serialize()).Info("Successfully sent NOTIFY packet.")
}
