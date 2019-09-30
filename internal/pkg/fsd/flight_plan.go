package fsd

import "github.com/pkg/errors"

// FlightPlan is the packet sent to update a pilot's flight plan
type FlightPlan struct {
	Callsign               string
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

// Parse parses a PLAN packet from FSD
func (f *FlightPlan) Parse(split []string) error {
	if len(split) >= 22 {
		f.Callsign = split[5]
		f.Type = split[7]
		f.Aircraft = split[8]
		f.CruiseSpeed = split[9]
		f.DepartureAirport = split[10]
		f.EstimatedDepartureTime = split[11]
		f.ActualDepartureTime = split[12]
		f.Altitude = split[13]
		f.DestinationAirport = split[14]
		f.HoursEnroute = split[15]
		f.MinutesEnroute = split[16]
		f.HoursFuel = split[17]
		f.MinutesFuel = split[18]
		f.AlternateAirport = split[19]
		f.Remarks = split[20]
		f.Route = split[21]
		return nil
	} else {
		return errors.Errorf("Invalid flight plan packet. +%v", reassemble(split))
	}
}
