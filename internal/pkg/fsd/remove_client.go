package fsd

import (
	"github.com/pkg/errors"
)

// RemoveClient is the packet sent to remove a client from the network
type RemoveClient struct {
	Callsign string
	Server   string
}

// Parse parses a RMCLIENT packet from FSD
func (r *RemoveClient) Parse(split []string) error {
	if len(split) >= 6 {
		r.Server = split[2]
		r.Callsign = split[5]
		return nil
	}
	return errors.Errorf("Invalid remove client packet. +%v", reassemble(split))
}
