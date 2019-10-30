package fsd

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// ATCData AD
type ATCData struct {
	Base
	Callsign     string
	Frequency    int
	FacilityType int
	VisualRange  int
	Rating       int
	Latitude     float64
	Longitude    float64
}

// Serialize converts a struct into an FSD packet
func (a ATCData) Serialize() string {
	msg := strings.Builder{}
	msg.WriteString("AD")
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
	msg.WriteString(a.Callsign)
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.Frequency))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.FacilityType))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.VisualRange))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(a.Rating))
	msg.WriteString(":")
	msg.WriteString(fmt.Sprintf("%f", a.Latitude))
	msg.WriteString(":")
	msg.WriteString(fmt.Sprintf("%f", a.Longitude))
	msg.WriteString(":")
	msg.WriteString("0") // Transceiver altitude
	return msg.String()
}

// DeserializeATCData maps an array of strings to an AddClient struct
func DeserializeATCData(fields []string) (ATCData, error) {
	if len(fields) >= 13 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		frequency, err := strconv.Atoi(fields[6])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse frequency. %v", reassemble(fields))
		}
		facilityType, err := strconv.Atoi(fields[7])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse facility type. %v", reassemble(fields))
		}
		visualRange, err := strconv.Atoi(fields[8])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse visual range. %v", reassemble(fields))
		}
		rating, err := strconv.Atoi(fields[9])
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse rating. %v", reassemble(fields))
		}
		latitude, err := strconv.ParseFloat(fields[10], 64)
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to parse latitude. %v", reassemble(fields))
		}
		longitude, err := strconv.ParseFloat(fields[11], 64)
		if err != nil {
			return ATCData{}, errors.Wrapf(err, "Failed to get longitude. %v", reassemble(fields))
		}
		return ATCData{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			Callsign:     fields[5],
			Frequency:    frequency,
			FacilityType: facilityType,
			VisualRange:  visualRange,
			Rating:       rating,
			Latitude:     latitude,
			Longitude:    longitude,
		}, nil
	}
	return ATCData{}, errors.Errorf("Invalid packet. %v", reassemble(fields))
}
