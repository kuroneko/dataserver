package dataserver

import (
	"dataserver/internal/pkg/fsd"
	"net/textproto"
)

// HandlePing deals with others servers checking if we are alive by ping
func HandlePing(fields []string, conn *textproto.Conn) error {
	ping, err := fsd.DeserializePing(fields)
	if err != nil {
		return err
	}
	sendPong(ping, conn)
	return nil
}
