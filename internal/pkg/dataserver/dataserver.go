package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/fsd"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io"
	"io/ioutil"
	"time"
)

// KafkaPayload used to format data sent to kafka
type KafkaPayload struct {
	MessageType string      `json:"message_type"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Channel streams the clientList updates
var (
	Channel = make(chan ClientList)

	totalConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dataserver_connections_total",
		Help: "The total number of FSD connections.",
	}, []string{"server"})

	packetsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dataserver_packets_processed",
		Help: "The total number of processed packets.",
	})

	timeToProcessPacket = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "dataserver_time_to_process_packet",
		Help: "The time to process a packet sent by FSD.",
	})
)

// AddFSDClient handles the creation of an FSD client to request data with
func (c *Context) AddFSDClient() {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.Fatal("Data server name not defined.")
	}

	// Initial setup
	c.sendAddClient(name)
	time.Sleep(time.Second)
	c.sendATCData(name)

	// Continually send updates every 30 seconds to keep the connection alive
	for range time.Tick(30 * time.Second) {
		c.sendAddClient(name)
		time.Sleep(time.Second)
		c.sendATCData(name)
	}
}

// EncodeJSON encodes the current Client list to JSON.
func EncodeJSON(clientList ClientList) ([]byte, error) {
	clientJSON, err := json.Marshal(clientList)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return clientJSON, nil
}

// RemoveTimedOutClients loops through the client lists to find clients who are no longer sending updates, and removes them.
func (c *Context) RemoveTimedOutClients() {
	time.Sleep(1 * time.Minute)
	for range time.Tick(10 * time.Second) {
		c.checkForTimeouts()
	}
}

// RequestATIS sends requests to all ATC clients for their ATIS every minute
func (c *Context) RequestATIS() {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.Fatal("Data server name not defined.")
	}

	// Initial setup, give the server 5 seconds to process the backlog of added clients
	time.Sleep(5 * time.Second)
	c.sendATISRequest(name)

	// Continue to request ATIS data every minute
	for range time.Tick(time.Minute) {
		c.sendATISRequest(name)
	}
}

// SetupServer handles the creation of an FSD server
func (c *Context) SetupServer() {
	// Initial setup
	c.SendNotify()
	time.Sleep(time.Second)
	c.SendSync()

	// Do this every 2 minutes from now on
	for range time.Tick(2 * time.Minute) {
		c.SendNotify()
		time.Sleep(time.Second)
		c.SendSync()
	}
}

// WriteDataFile overwrites the data file with new data.
func WriteDataFile(clientJSON []byte) error {
	directory, err := config.Cfg.String("data.file.directory")
	if err != nil {
		log.Fatal("Data file directory not defined.")
	}
	err = ioutil.WriteFile(directory+"vatsim-data.json", clientJSON, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// checkForTimeouts loops through all clients and checks if they have timed out
func (c *Context) checkForTimeouts() {
	c.ClientList.Mutex.Lock()
	defer c.ClientList.Mutex.Unlock()
	for i := 0; i < len(c.ClientList.ATCData); i++ {
		if time.Since(*&c.ClientList.ATCData[i].LastUpdated) >= (30 * time.Second) {
			kafkaPush(c.Producer, *&c.ClientList.ATCData[i], "remove_client")
			totalConnections.With(prometheus.Labels{"server": *&c.ClientList.ATCData[i].Server}).Dec()
			log.WithFields(log.Fields{
				"callsign": *&c.ClientList.ATCData[i].Callsign,
				"server":   *&c.ClientList.ATCData[i].Server,
			}).Info("Client timed out.")
			*&c.ClientList.ATCData = append(c.ClientList.ATCData[:i], c.ClientList.ATCData[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(c.ClientList.PilotData); i++ {
		if time.Since(*&c.ClientList.PilotData[i].LastUpdated) >= (30 * time.Second) {
			kafkaPush(c.Producer, *&c.ClientList.PilotData[i], "remove_client")
			totalConnections.With(prometheus.Labels{"server": *&c.ClientList.PilotData[i].Server}).Dec()
			log.WithFields(log.Fields{
				"callsign": *&c.ClientList.PilotData[i].Callsign,
				"server":   *&c.ClientList.PilotData[i].Server,
			}).Info("Client timed out.")
			*&c.ClientList.PilotData = append(c.ClientList.PilotData[:i], c.ClientList.PilotData[i+1:]...)
			i--
		}
	}
}

// kafkaPush publishes to the Kafka feed
func kafkaPush(producer *kafka.Producer, data interface{}, messageType string) {
	topic := "datafeed"
	kafkaData := KafkaPayload{
		MessageType: messageType,
		Data:        data,
		Timestamp:   time.Now().UTC(),
	}
	jsonData, _ := json.Marshal(kafkaData)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          jsonData,
	}, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"payload": kafkaData,
			"error":   err,
		}).Error("Failed to send Kafka payload.")
	}
}

// Listen continually reads, parses and handles FSD packets.
func (c *Context) Listen() {
	for {
		bytes, err := fsd.ReadMessage(c.Consumer)
		timer := prometheus.NewTimer(timeToProcessPacket)
		if err == io.EOF {
			log.Fatal("FSD connection close.")
		} else if err != nil {
			log.WithField("error", err).Error("Failed to read message from FSD connection.")
		}
		fields := fsd.ParseMessage(bytes)
		c.processMessage(fields)
		timer.ObserveDuration()
	}
}

// processMessage classifies the FSD packet and performs the appropriate action
func (c *Context) processMessage(fields []string) {
	switch fields[0] {
	case "ADDCLIENT":
		err := c.HandleAddClient(fields)
		checkPacketHandlerError(err, "ADDCLIENT")
		break
	case "RMCLIENT":
		err := c.RemoveClient(fields)
		checkPacketHandlerError(err, "RMCLIENT")
		break
	case "PD":
		err := c.HandlePilotData(fields)
		checkPacketHandlerError(err, "PD")
		break
	case "AD":
		err := c.HandleATCData(fields)
		checkPacketHandlerError(err, "AD")
		break
	case "PLAN":
		err := c.HandleFlightPlan(fields)
		checkPacketHandlerError(err, "PLAN")
		break
	case "PING":
		err := c.HandlePing(fields)
		checkPacketHandlerError(err, "PING")
		break
	case "MC":
		if fields[5] == "25" {
			err := c.HandleATISData(fields)
			checkPacketHandlerError(err, "MC")
		}
		break
	}
	packetsProcessed.Inc()
}

func checkPacketHandlerError(err error, packet string) {
	if err != nil {
		log.WithField("error", err).Errorf("Failed to handle %s packet.", packet)
	}
}
