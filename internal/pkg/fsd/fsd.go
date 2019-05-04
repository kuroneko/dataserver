package fsd

import (
	"bufio"
	"dataserver/internal/pkg/dataserver"
	"net"
	"strconv"
	"strings"
)

// pdCount is a count of the number of packets we have sent to the FSD server.
var pdCount int

// Connect establishes a connection to the FSD server.
func Connect() net.Conn {
	ip, err := dataserver.Cfg.String("fsd.server.ip")
	if err != nil {
		panic(err)
	}
	port, err := dataserver.Cfg.String("fsd.server.port")
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		panic(err)
	}
	return conn
}

// send formats and sends a new FSD packet to the FSD server.
func send(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message + "\r\n"))
	if err != nil {
		panic(err)
	}
	pdCount++
}

// ParseMessage splits an FSD message based on the colon delimiter for further handling.
func ParseMessage(bytes []byte) []string {
	split := strings.Split(strings.Trim(string(bytes), " \r\n"), ":")
	return split
}

// ReadMessage reads a new FSD message from the buffer reader.
func ReadMessage(bufReader *bufio.Reader) []byte {
	bytes, err := bufReader.ReadBytes('\n')
	if err != nil {
		panic(err)
	}
	return bytes
}

// Sync sends a sync packet to the FSD server.
func Sync(conn net.Conn) {
	name, err := dataserver.Cfg.String("data.server.name")
	if err != nil {
		panic(err)
	}
	send(conn, "SYNC:*:"+name+":B1:1:")
}

// Pong returns an FSD server's ping request.
func Pong(conn net.Conn, split []string) {
	name, err := dataserver.Cfg.String("data.server.name")
	if err != nil {
		panic(err)
	}
	send(conn, "PONG:"+split[2]+":"+name+":U"+strconv.Itoa(pdCount)+":1"+split[5])
}

// SetupReader wraps the FSD connection in a buffered reader for easier ingestion.
func SetupReader(conn net.Conn) *bufio.Reader {
	bufReader := bufio.NewReader(conn)
	return bufReader
}
