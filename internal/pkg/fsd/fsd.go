package fsd

import (
	"dataserver/internal/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/textproto"
	"strings"
)

// PdCount is a count of the number of packets we have sent to the FSD server.
var PdCount int

// Connect establishes a connection to the FSD server.
func Connect() *textproto.Conn {
	ip, err := config.Cfg.String("fsd.server.ip")
	if err != nil {
		log.Fatal("FSD IP not defined.")
	}
	port, err := config.Cfg.String("fsd.server.port")
	if err != nil {
		log.Fatal("FSD port not defined.")
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

// Send formats and sends a new FSD packet to the FSD server.
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
