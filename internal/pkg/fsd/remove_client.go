package fsd

import (
	"github.com/pkg/errors"
)

// RemoveClient is the packet sent to remove a client from the network
type RemoveClient struct {
	Callsign string
}

// Parse parses a RMCLIENT packet from FSD
func (r *RemoveClient) Parse(split []string) error {
	if len(split) >= 6 {
		r.Callsign = split[5]
		return nil
	} else {
		return errors.Errorf("Invalid remove client packet. +%v", reassemble(split))
	}
}
