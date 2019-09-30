package dataserver

import (
	"bufio"
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"fmt"
	"github.com/getsentry/sentry-go"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net"
	"time"
	
)

// Start connects and begins parsing and saving data files.
func Start() {
	config.ReadConfig()
	config.ConfigureSentry()
	conn := fsd.Connect()
	defer func() {
		if err := conn.Close(); err != nil {
			sentry.CaptureException(err)
		}
	}()
	bufReader := fsd.SetupReader(conn)
	fsd.Sync(conn)
	clientList := &dataserver.ClientList{}
	go update()
	listen(bufReader, clientList, conn)
}

// update handles the creation of a 15 second ticker for updating the data file
func update() {
	now := time.Now().UTC()
	for clientList := range dataserver.Channel {
		if time.Since(now) >= (15 * time.Second) {
			err := updateFile(clientList)
			if err != nil {
				sentry.CaptureException(err)
			}
			now = time.Now().UTC()
		}
	}
}

// listen continually reads, parses and handles FSD packets.
func listen(bufReader *bufio.Reader, clientList *dataserver.ClientList, conn net.Conn) {
	fmt.Printf("%+v Starting Kafka connection\n", time.Now().UTC().Format(time.RFC3339))

	kafkaServer, err := dataserver.Cfg.String("kafka.server")
	if err != nil {
		fmt.Printf("%+v Kafka server not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaUsername, err := dataserver.Cfg.String("kafka.credentials.username")
	if err != nil {
		fmt.Printf("%+v Kafka username not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}	
	kafkaPassword, err := dataserver.Cfg.String("kafka.credentials.password")
	if err != nil {
		fmt.Printf("%+v Kafka password not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaProtocol, err := dataserver.Cfg.String("kafka.credentials.protocol")
	if err != nil {
		fmt.Printf("%+v Kafka protocol not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaMechanism, err := dataserver.Cfg.String("kafka.credentials.mechanism")
	if err != nil {
		fmt.Printf("%+v Kafka authentication mechanism not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
		"sasl.username": kafkaUsername,
		"sasl.password": kafkaPassword,
		"security.protocol": kafkaProtocol,
		"sasl.mechanism": kafkaMechanism,

	})

	if err != nil {
		fmt.Printf("%+v Kafka connection failed\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	fmt.Printf("%+v Kafka connected\n", time.Now().UTC().Format(time.RFC3339))
	defer producer.Close()
	for {
		bytes, err := fsd.ReadMessage(bufReader)
		if err != nil {
			sentry.CaptureException(err)
			continue
		}
		split := fsd.ParseMessage(bytes)
		processMessage(split, clientList, conn, producer)
	}
}

// processMessage classifies the FSD packet and performs the appropriate action
func processMessage(split []string, clientList *dataserver.ClientList, conn net.Conn, producer *kafka.Producer) {
	if split[0] == "ADDCLIENT" {
		checkError(dataserver.AddClient(split, clientList, producer))
	} else if split[0] == "RMCLIENT" {
		checkError(dataserver.RemoveClient(split, clientList, producer))
	} else if split[0] == "PD" {
		checkError(dataserver.UpdatePosition(split, clientList, producer))
	} else if split[0] == "AD" {
		checkError(dataserver.UpdateControllerData(split, clientList, producer))
	} else if split[0] == "PLAN" {
		checkError(dataserver.UpdateFlightPlan(split, clientList, producer))
	} else if split[0] == "PING" && len(split) >= 6 {
		fsd.Pong(conn, split)
	}
}

// checkError checks if an error occurred and reports it to Sentry
func checkError(err error) {
	if err != nil {
		sentry.CaptureException(err)
	}
}

// updateFile encodes the current clientList and prints to the data file
func updateFile(clientList dataserver.ClientList) error {
	clientJSON, err := dataserver.EncodeJSON(clientList)
	if err != nil {
		return err
	}
	err = dataserver.WriteDataFile(clientJSON)
	if err != nil {
		return err
	}
	fmt.Printf("%+v Data file updated\n", time.Now().UTC().Format(time.RFC3339))
	return nil
}
