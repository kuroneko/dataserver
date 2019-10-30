package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// FlightPlan describes the data about a filed flight plan.
type FlightPlan struct {
	FlightRules string         `json:"flight_rules"`
	Aircraft    string         `json:"aircraft"`
	CruiseSpeed string         `json:"cruise_speed"`
	Departure   string         `json:"departure"`
	Arrival     string         `json:"arrival"`
	Altitude    string         `json:"altitude"`
	Alternate   string         `json:"alternate"`
	Route       string         `json:"route"`
	Time        FlightPlanTime `json:"time"`
	Remarks     string         `json:"remarks"`
}

// FlightPlanTime represents the times present in a filed flight plan.
type FlightPlanTime struct {
	Departure      string `json:"departure"`
	HoursEnroute   string `json:"hours_enroute"`
	MinutesEnroute string `json:"minutes_enroute"`
	HoursFuel      string `json:"hours_fuel"`
	MinutesFuel    string `json:"minutes_fuel"`
}

// HandleFlightPlan updates the flight plan entry for the specified callsign
func HandleFlightPlan(fields []string, clientList *ClientList, producer *kafka.Producer) error {
	flightPlan, err := fsd.DeserializeFlightPlan(fields)
	if err != nil {
		return err
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == flightPlan.Callsign {
			*&clientList.PilotData[i].FlightPlan = FlightPlan{
				FlightRules: flightPlan.Type,
				Aircraft:    flightPlan.Aircraft,
				CruiseSpeed: flightPlan.CruiseSpeed,
				Departure:   flightPlan.DepartureAirport,
				Altitude:    flightPlan.Altitude,
				Arrival:     flightPlan.DestinationAirport,
				Alternate:   flightPlan.AlternateAirport,
				Remarks:     flightPlan.Remarks,
				Route:       flightPlan.Route,
				Time: FlightPlanTime{
					Departure:      flightPlan.EstimatedDepartureTime,
					HoursEnroute:   flightPlan.HoursEnroute,
					MinutesEnroute: flightPlan.MinutesEnroute,
					HoursFuel:      flightPlan.HoursFuel,
					MinutesFuel:    flightPlan.MinutesFuel,
				},
			}
			kafkaPush(producer, clientList.PilotData[i].FlightPlan, "update_flight_plan")
			break
		}
	}
	log.WithFields(log.Fields{
		"callsign":  flightPlan.Callsign,
		"aircraft":  flightPlan.Aircraft,
		"departure": flightPlan.DepartureAirport,
		"arrival":   flightPlan.DestinationAirport,
		"altitude":  flightPlan.Altitude,
	}).Info("Flight plan packet received.")
	Channel <- *clientList
	return nil
}
