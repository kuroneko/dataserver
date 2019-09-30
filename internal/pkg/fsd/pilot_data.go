package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// PilotData is the packet sent to update data about a pilot
type PilotData struct {
	Callsign    string
	Latitude    float64
	Longitude   float64
	Altitude    int
	GroundSpeed int
	Heading     int
}

// Parse parses a PD packet from FSD
func (p *PilotData) Parse(split []string) error {
	if len(split) >= 13 {
		latitude, err := strconv.ParseFloat(split[9], 64)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse latitude. %+v", reassemble(split))
		}
		longitude, err := strconv.ParseFloat(split[10], 64)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse longitude. %+v", reassemble(split))
		}
		altitude, err := strconv.Atoi(split[11])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse altitude. %+v", reassemble(split))
		}
		speed, err := strconv.Atoi(split[12])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse speed. %+v", reassemble(split))
		}
		heading, err := getHeading(split[13])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse heading. %+v", reassemble(split))
		}
		p.Callsign = split[6]
		p.Latitude = latitude
		p.Longitude = longitude
		p.Altitude = altitude
		p.GroundSpeed = speed
		p.Heading = heading
		return nil
	}
	return errors.Errorf("Invalid pilot data packet. +%v", reassemble(split))
}

// getHeading parses the PBH FSD value to extract the heading
func getHeading(split string) (int, error) {
	pbh, err := strconv.ParseUint(split, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse PBH field %+v", split)
	}
	hdgBit := (pbh >> 2) & 0x3FF
	heading := float64(hdgBit) / 1024.0 * 360
	if heading < 0.0 {
		heading += 360
	} else if heading >= 360.0 {
		heading -= 360
	}
	return int(heading), nil
}
