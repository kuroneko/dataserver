package dataserver

import (
	"dataserver/internal/pkg/fsd"
)

// HandlePing deals with others servers checking if we are alive by ping
func (c *Context) HandlePing(fields []string) error {
	ping, err := fsd.DeserializePing(fields)
	if err != nil {
		return err
	}
	c.sendPong(ping)
	return nil
}
