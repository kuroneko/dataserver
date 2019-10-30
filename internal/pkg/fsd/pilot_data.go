package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// PilotData PD
type PilotData struct {
	Base
	IdentFlag   string
	Callsign    string
	Transponder int
	Rating      int
	Latitude    float64
	Longitude   float64
	Altitude    int
	GroundSpeed int
	Heading     int
}

// DeserializePilotData maps an array of strings to a PilotData struct
func DeserializePilotData(fields []string) (PilotData, error) {
	if len(fields) >= 14 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		transponder, err := strconv.Atoi(fields[7])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse transponder. %v", reassemble(fields))
		}
		rating, err := strconv.Atoi(fields[8])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse rating. %v", reassemble(fields))
		}
		latitude, err := strconv.ParseFloat(fields[9], 64)
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse latitude. %v", reassemble(fields))
		}
		longitude, err := strconv.ParseFloat(fields[10], 64)
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse longitude. %v", reassemble(fields))
		}
		altitude, err := strconv.Atoi(fields[11])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse altitude. %v", reassemble(fields))
		}
		speed, err := strconv.Atoi(fields[12])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse speed. %v", reassemble(fields))
		}
		heading, err := getHeading(fields[13])
		if err != nil {
			return PilotData{}, errors.Wrapf(err, "Failed to parse heading. %v", reassemble(fields))
		}
		return PilotData{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			IdentFlag:   fields[5],
			Callsign:    fields[6],
			Transponder: transponder,
			Rating:      rating,
			Latitude:    latitude,
			Longitude:   longitude,
			Altitude:    altitude,
			GroundSpeed: speed,
			Heading:     heading,
		}, nil
	}
	return PilotData{}, errors.Errorf("Invalid packet. %v", reassemble(fields))
}

// getHeading parses the PBH FSD value to extract the heading
func getHeading(fields string) (int, error) {
	pbh, err := strconv.ParseUint(fields, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse PBH field %+v", fields)
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
