package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// ATCData is the packet sent to update data about a controller
type ATCData struct {
	Callsign     string
	Frequency    int
	FacilityType int
	VisualRange  int
	Latitude     float64
	Longitude    float64
}

// Parse parses an AD packet from FSD
func (a *ATCData) Parse(split []string) error {
	if len(split) >= 12 {
		frequency, err := strconv.Atoi(split[6])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse frequency. %+v", reassemble(split))
		}
		facilityType, err := strconv.Atoi(split[7])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse facility type. %+v", reassemble(split))
		}
		visualRange, err := strconv.Atoi(split[8])
		if err != nil {
			return errors.Wrapf(err, "Failed to parse visual range. %+v", reassemble(split))
		}
		latitude, err := strconv.ParseFloat(split[10], 64)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse latitude. %+v", reassemble(split))
		}
		longitude, err := strconv.ParseFloat(split[11], 64)
		if err != nil {
			return errors.Wrapf(err, "Failed to get longitude. %+v", reassemble(split))
		}
		a.Callsign = split[5]
		a.Frequency = frequency
		a.FacilityType = facilityType
		a.VisualRange = visualRange
		a.Latitude = latitude
		a.Longitude = longitude
		return nil
	} else {
		return errors.Errorf("Invalid ATC data packet. +%v", reassemble(split))
	}
}
