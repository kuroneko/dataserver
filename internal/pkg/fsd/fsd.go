package fsd

import (
	"dataserver/internal/pkg/config"
	"github.com/pkg/errors"
	"net/textproto"
	"strconv"
	"strings"
)

// pdCount is a count of the number of packets we have sent to the FSD server.
var pdCount int

// Connect establishes a connection to the FSD server.
func Connect() *textproto.Conn {
	ip, err := config.Cfg.String("fsd.server.ip")
	if err != nil {
		panic(err)
	}
	port, err := config.Cfg.String("fsd.server.port")
	if err != nil {
		panic(err)
	}
	conn, err := textproto.Dial("tcp", ip+":"+port)
	if err != nil {
		panic(err)
	}
	return conn
}

// send formats and sends a new FSD packet to the FSD server.
func send(conn *textproto.Conn, message string) error {
	_, err := conn.Cmd(message)
	if err != nil {
		return errors.Wrapf(err, "Failed to send packet to FSD server. %+v", conn)
	}
	pdCount++
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
		panic(err)
	}
	err = send(conn, "SYNC:*:"+name+":B1:1:")
	if err != nil {
		panic(err)
	}
}

// Pong returns an FSD server's ping request.
func Pong(conn *textproto.Conn, split []string) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		panic(err)
	}
	err = send(conn, "PONG:"+split[2]+":"+name+":U"+strconv.Itoa(pdCount)+":1"+split[5])
	if err != nil {
		panic(err)
	}
}

// reassemble puts the FSD packet back together for debugging
func reassemble(split []string) string {
	return strings.Join(split, ":")
}
