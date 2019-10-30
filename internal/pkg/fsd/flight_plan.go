package fsd

import (
	"github.com/pkg/errors"
	"strconv"
)

// FlightPlan PLAN
type FlightPlan struct {
	Base
	Callsign               string
	Revision               string
	Type                   string
	Aircraft               string
	CruiseSpeed            string
	DepartureAirport       string
	EstimatedDepartureTime string
	ActualDepartureTime    string
	Altitude               string
	DestinationAirport     string
	HoursEnroute           string
	MinutesEnroute         string
	HoursFuel              string
	MinutesFuel            string
	AlternateAirport       string
	Remarks                string
	Route                  string
}

// DeserializeFlightPlan maps an array of strings to a FlightPlan struct
func DeserializeFlightPlan(fields []string) (FlightPlan, error) {
	if len(fields) >= 22 {
		packetNumber, err := strconv.Atoi(fields[3][1:])
		if err != nil {
			return FlightPlan{}, errors.Wrapf(err, "Failed to parse packet number. %v", reassemble(fields))
		}
		hopCount, err := strconv.Atoi(fields[4])
		if err != nil {
			return FlightPlan{}, errors.Wrapf(err, "Failed to parse hop count. %v", reassemble(fields))
		}
		return FlightPlan{
			Base: Base{
				Destination:  fields[1],
				Source:       fields[2],
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			Callsign:               fields[5],
			Revision:               fields[6],
			Type:                   fields[7],
			Aircraft:               fields[8],
			CruiseSpeed:            fields[9],
			DepartureAirport:       fields[10],
			EstimatedDepartureTime: fields[11],
			ActualDepartureTime:    fields[12],
			Altitude:               fields[13],
			DestinationAirport:     fields[14],
			HoursEnroute:           fields[15],
			MinutesEnroute:         fields[16],
			HoursFuel:              fields[17],
			MinutesFuel:            fields[18],
			AlternateAirport:       fields[19],
			Remarks:                fields[20],
			Route:                  fields[21],
		}, nil
	}
	return FlightPlan{}, errors.Errorf("Invalid packet. %v", reassemble(fields))
}
