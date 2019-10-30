package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// ATISData MC 25
type ATISData struct {
	Base
	From string
	Type string
	Data string
}

// DeserializeATISData maps an array of strings to an ATISData struct
func DeserializeATISData(fields []string) (ATISData, error) {
	if len(fields) >= 10 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return ATISData{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return ATISData{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		return ATISData{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			From: fields[6],
			Type: fields[8],
			Data: fields[9],
		}, nil
	}
	return ATISData{}, errors.Errorf("Invalid remove client packet. %v", reassemble(fields))
}
