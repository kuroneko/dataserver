package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// Ping PING
type Ping struct {
	Base
	Data string
}

// DeserializePing maps an array of strings to a Ping struct
func DeserializePing(fields []string) (Ping, error) {
	if len(fields) >= 6 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return Ping{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return Ping{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		return Ping{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			Data: fields[5],
		}, nil
	}
	return Ping{}, errors.Errorf("Invalid packet. %v", reassemble(fields))
}
