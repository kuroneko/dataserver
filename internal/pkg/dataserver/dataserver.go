package dataserver

import (
	"dataserver/internal/pkg/config"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
	"net/textproto"
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
)

// kafkaPush publishes to the Kafka feed
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
		Value:          jsonData,
	}, nil)
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

// RequestATIS sends requests to all ATC clients for their ATIS every minute
func RequestATIS(conn *textproto.Conn, clientList *ClientList) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		panic(err)
	}

	// Initial setup, give the server 5 seconds to process the backlog of added clients
	time.Sleep(5 * time.Second)
	sendATISRequest(clientList, name, conn)

	// Continue to request ATIS data every minute
	for range time.Tick(time.Minute) {
		sendATISRequest(clientList, name, conn)
	}
}

// AddFSDClient handles the creation of an FSD client to request data with
func AddFSDClient(conn *textproto.Conn) {
	name, err := config.Cfg.String("data.server.name")
	if err != nil {
		log.WithFields(log.Fields{
			"config": config.Cfg,
			"error":  err,
		}).Panic("Failed to get server name from configuration.")
	}

	// Initial setup
	sendAddClient(name, err, conn)
	time.Sleep(time.Second)
	sendATCData(name, err, conn)

	// Continually send updates every 30 seconds to keep the connection alive
	for range time.Tick(30 * time.Second) {
		sendAddClient(name, err, conn)
		time.Sleep(time.Second)
		sendATCData(name, err, conn)
	}
}
