package fsd

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// ATISRequest MC 24
type ATISRequest struct {
	Base
	From string
}

// Serialize converts a struct into an FSD packet
func (a ATISRequest) Serialize() string {
	msg := strings.Builder{}
	msg.WriteString("MC")
	msg.WriteString(":")
	msg.WriteString("%%")
	msg.WriteString(a.Destination)
	msg.WriteString(":")
	msg.WriteString(a.Source)
	msg.WriteString(":")
	msg.WriteString("U")
	msg.WriteString(strconv.Itoa(a.PacketNumber))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.HopCount))
	msg.WriteString(":")
	msg.WriteString("24")
	msg.WriteString(":")
	msg.WriteString(a.From)
	msg.WriteString(":")
	msg.WriteString("ATIS")
	return msg.String()
}

// DeserializeATISRequest maps an array of strings to an ATISRequest struct
func DeserializeATISRequest(fields []string) (ATISRequest, error) {
	if len(fields) >= 8 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return ATISRequest{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return ATISRequest{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		return ATISRequest{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			From: fields[6],
		}, nil
	}
	return ATISRequest{}, errors.Errorf("Invalid remove client packet. %v", reassemble(fields))
}
