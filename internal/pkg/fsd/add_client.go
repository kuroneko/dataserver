package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// AddClient is the packet sent to add a client to the network
type AddClient struct {
	CID      int
	Server   string
	Callsign string
	Type     int
	Rating   int
	RealName string
}

// Parse parses an ADDCLIENT packet from FSD
func (a *AddClient) Parse(split []string) error {
	if len(split) >= 12 {
		cid, err := strconv.Atoi(split[5])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse CID. %+v", reassemble(split))
		}
		rating, err := strconv.Atoi(split[9])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse rating. %+v", reassemble(split))
		}
		clientType, err := strconv.Atoi(split[8])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse client type. %+v", reassemble(split))
		}
		a.CID = cid
		a.Server = split[6]
		a.Callsign = split[7]
		a.Type = clientType
		a.Rating = rating
		a.RealName = split[11]
		return nil
	}
	return errors.Errorf("Invalid add client packet. +%v", reassemble(split))
}
