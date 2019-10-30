package fsd

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// AddClient ADDCLIENT
type AddClient struct {
	Base
	CID              int
	Server           string
	Callsign         string
	Type             int
	Rating           int
	ProtocolRevision int
	RealName         string
	SimType          int
	Hidden           int
}

// Serialize converts a struct into an FSD packet
func (a AddClient) Serialize() string {
	msg := strings.Builder{}
	msg.WriteString("ADDCLIENT")
	msg.WriteString(":")
	msg.WriteString(a.Destination)
	msg.WriteString(":")
	msg.WriteString(a.Source)
	msg.WriteString(":")
	msg.WriteString("B")
	msg.WriteString(strconv.Itoa(a.PacketNumber))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.HopCount))
	msg.WriteString(":")
	if a.CID != 0 {
		msg.WriteString(strconv.Itoa(a.CID))
	}
	msg.WriteString(":")
	msg.WriteString(a.Server)
	msg.WriteString(":")
	msg.WriteString(a.Callsign)
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.Type))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.Rating))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.ProtocolRevision))
	msg.WriteString(":")
	msg.WriteString(a.RealName)
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.SimType))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.Hidden))
	return msg.String()
}

// DeserializeAddClient maps an array of strings to an AddClient struct
func DeserializeAddClient(fields []string) (AddClient, error) {
	if len(fields) >= 12 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		cid, err := strconv.Atoi(fields[5])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse CID. %v", reassemble(fields))
		}
		rating, err := strconv.Atoi(fields[9])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse rating. %v", reassemble(fields))
		}
		clientType, err := strconv.Atoi(fields[8])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse client type. %v", reassemble(fields))
		}
		protocolRevision, err := strconv.Atoi(fields[10])
		if err != nil {
			return AddClient{}, errors.Wrapf(err, "Failed to parse protocol revision. %v", reassemble(fields))
		}
		if len(fields) > 12 {
			simType, err := strconv.Atoi(fields[12])
			if err != nil {
				return AddClient{}, errors.Wrapf(err, "Failed to parse simulator type. %v", reassemble(fields))
			}
			hidden, err := strconv.Atoi(fields[13])
			if err != nil {
				return AddClient{}, errors.Wrapf(err, "Failed to parse hidden flag. %v", reassemble(fields))
			}
			return AddClient{
				Base: Base{
					Destination:  fields[1],
					Source:       fields[2],
					PacketNumber: packetNumber,
					HopCount:     hopCount,
				},
				CID:              cid,
				Server:           fields[6],
				Callsign:         fields[7],
				Type:             clientType,
				Rating:           rating,
				ProtocolRevision: protocolRevision,
				RealName:         fields[11],
				SimType:          simType,
				Hidden:           hidden,
			}, nil
		}
		return AddClient{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			CID:              cid,
			Server:           fields[6],
			Callsign:         fields[7],
			Type:             clientType,
			Rating:           rating,
			ProtocolRevision: protocolRevision,
			RealName:         fields[11],
		}, nil
	}
	return AddClient{}, errors.Errorf("Invalid packet. %v", reassemble(fields))
}
