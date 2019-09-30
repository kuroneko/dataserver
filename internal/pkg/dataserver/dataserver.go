package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/fsd"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
	"time"
)

// ClientList is a list of all clients currently connected to the network.
type ClientList struct {
	PilotData []PilotData `json:"pilots"`
	ATCData   []ATCData   `json:"controllers"`
}

// PilotData is data about individual pilots on the network.
type PilotData struct {
	Server     string     `json:"server"`
	Callsign   string     `json:"callsign"`
	Member     MemberData `json:"member"`
	Latitude   float64    `json:"latitude"`
	Longitude  float64    `json:"longitude"`
	Altitude   int        `json:"altitude"`
	Speed      int        `json:"speed"`
	Heading    int        `json:"heading"`
	FlightPlan FlightPlan `json:"plan"`
}

// MemberData represents a user's personal data.
type MemberData struct {
	CID  int    `json:"cid"`
	Name string `json:"name"`
}

// ATCData is data about individual controllers on the network.
type ATCData struct {
	Server       string     `json:"server"`
	Callsign     string     `json:"callsign"`
	Member       MemberData `json:"member"`
	Rating       int        `json:"rating"`
	Frequency    int        `json:"frequency"`
	FacilityType int        `json:"facility"`
	VisualRange  int        `json:"range"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
}

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

// KafkaPayload used to format data sent to kafka
type KafkaPayload struct {
	MessageType string      `json:"message_type"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Channel streams the clientList updates
var Channel = make(chan ClientList)

func kafkaPush(producer *kafka.Producer, data interface{}, messageType string) {
	topic := "datafeed"
	kafkaData := KafkaPayload{
		MessageType: messageType,
		Data:        data,
		Timestamp:   time.Now().UTC(),
	}
	jsonData, _ := json.Marshal(kafkaData)
	_ = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(jsonData),
	}, nil)
}

// UpdatePosition updates a client's position data in the Client list and updates the JSON file.
func UpdatePosition(split []string, clientList *ClientList, producer *kafka.Producer) error {
	pilotData := &fsd.PilotData{}
	err := pilotData.Parse(split)
	if err != nil {
		return err
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == split[6] {
			*&clientList.PilotData[i].Latitude = pilotData.Latitude
			*&clientList.PilotData[i].Longitude = pilotData.Longitude
			*&clientList.PilotData[i].Altitude = pilotData.Altitude
			*&clientList.PilotData[i].Speed = pilotData.GroundSpeed
			*&clientList.PilotData[i].Heading = pilotData.Heading
			kafkaPush(producer, clientList.PilotData[i], "update_position")
			break
		}
	}
	fmt.Printf("%+v Pilot Data Received: %+v\n", time.Now().UTC().Format(time.RFC3339), pilotData.Callsign)
	Channel <- *clientList
	return nil
}

// UpdateFlightPlan updates the flight plan entry for the specified callsign
func UpdateFlightPlan(split []string, clientList *ClientList, producer *kafka.Producer) error {
	flightPlan := &fsd.FlightPlan{}
	err := flightPlan.Parse(split)
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
	fmt.Printf("%+v Flight Plan Update Received: %+v\n", time.Now().UTC().Format(time.RFC3339), flightPlan.Callsign)
	Channel <- *clientList
	return nil
}

// UpdateControllerData updates a controllers's data in the Client list and updates the JSON file.
func UpdateControllerData(split []string, clientList *ClientList, producer *kafka.Producer) error {
	atcData := &fsd.ATCData{}
	err := atcData.Parse(split)
	if err != nil {
		return err
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == atcData.Callsign {
			*&clientList.ATCData[i].Frequency = atcData.Frequency
			*&clientList.ATCData[i].FacilityType = atcData.FacilityType
			*&clientList.ATCData[i].VisualRange = atcData.VisualRange
			*&clientList.ATCData[i].Latitude = atcData.Latitude
			*&clientList.ATCData[i].Longitude = atcData.Longitude
			kafkaPush(producer, clientList.ATCData[i], "update_controller_data")
			break
		}
	}
	fmt.Printf("%+v ATC Data Received: %+v\n", time.Now().UTC().Format(time.RFC3339), atcData.Callsign)
	Channel <- *clientList
	return nil
}

// RemoveClient removes a client from the Client list and updates the JSON file.
func RemoveClient(split []string, clientList *ClientList, producer *kafka.Producer) error {
	removeClient := &fsd.RemoveClient{}
	err := removeClient.Parse(split)
	if err != nil {
		return err
	}
	for i, v := range clientList.PilotData {
		if v.Callsign == removeClient.Callsign {
			*&clientList.PilotData = append(clientList.PilotData[:i], clientList.PilotData[i+1:]...)
			data := PilotData{
				Server:   "",
				Callsign: removeClient.Callsign,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}

			kafkaPush(producer, data, "remove_client")
			break
		}
	}
	for i, v := range clientList.ATCData {
		if v.Callsign == removeClient.Callsign {
			*&clientList.ATCData = append(clientList.ATCData[:i], clientList.ATCData[i+1:]...)
			data := ATCData{
				Server:   "",
				Callsign: removeClient.Callsign,
				Rating:   0,
				Member: MemberData{
					CID:  0,
					Name: "",
				},
			}

			kafkaPush(producer, data, "remove_client")
			break
		}
	}
	fmt.Printf("%+v Client Deleted: %+v\n", time.Now().UTC().Format(time.RFC3339), removeClient.Callsign)
	Channel <- *clientList
	return nil
}

// AddClient adds a client to the Client list and updates the JSON file.
func AddClient(split []string, clientList *ClientList, producer *kafka.Producer) error {
	addClient := &fsd.AddClient{}
	err := addClient.Parse(split)
	if err != nil {
		return err
	}
	if addClient.Type == 1 {
		data := PilotData{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&clientList.PilotData = append(clientList.PilotData, data)
		kafkaPush(producer, data, "add_client")
	} else if addClient.Type == 2 {
		data := ATCData{
			Server:   addClient.Server,
			Callsign: addClient.Callsign,
			Rating:   addClient.Rating,
			Member: MemberData{
				CID:  addClient.CID,
				Name: addClient.RealName,
			},
		}
		*&clientList.ATCData = append(clientList.ATCData, data)
		kafkaPush(producer, data, "add_client")
	}
	fmt.Printf("%+v Client Added: %+v\n", time.Now().UTC().Format(time.RFC3339), addClient.Callsign)
	Channel <- *clientList
	return nil
}

// WriteDataFile overwrites the data file with new data.
func WriteDataFile(clientJSON []byte) error {
	directory, err := config.Cfg.String("data.file.directory")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(directory+"vatsim-data.json", clientJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file. %+v", clientJSON)
	}
	return nil
}

// EncodeJSON encodes the current Client list to JSON.
func EncodeJSON(clientList ClientList) ([]byte, error) {
	clientJSON, err := json.Marshal(clientList)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to encode client list to JSON. %+v", clientList)
	}
	return clientJSON, nil
}
