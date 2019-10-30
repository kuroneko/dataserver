package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// RemoveClient RMCLIENT
type RemoveClient struct {
	Base
	Callsign string
}

// DeserializeRemoveClient maps an array of strings to a RemoveClient struct
func DeserializeRemoveClient(fields []string) (RemoveClient, error) {
	if len(fields) >= 6 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return RemoveClient{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return RemoveClient{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		return RemoveClient{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			Callsign: fields[5],
		}, nil
	}
	return RemoveClient{}, errors.Errorf("Invalid remove client packet. %v", reassemble(fields))
}
