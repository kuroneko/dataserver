package config

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/olebedev/config"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
	"time"
)

// Cfg contains all of the necessary configuration data.
var Cfg *config.Config

// ReadConfig reads the config file and instantiates the config object
func ReadConfig() {
	file, err := ioutil.ReadFile("configs/config.yml")
	if err != nil {
		panic(err)
	}
	yamlString := string(file)
	Cfg, err = config.ParseYaml(yamlString)
	if err != nil {
		panic(err)
	}
}

// ConfigureSentry sets up Sentry for panic reporting
func ConfigureSentry() {
	dsn, err := Cfg.String("sentry.credentials.dsn")
	if err != nil {
		panic(err)
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		panic(err)
	}
}

// ConfigureKafka sets the Kafka connection up
func ConfigureKafka() *kafka.Producer {
	fmt.Printf("%+v Starting Kafka connection\n", time.Now().UTC().Format(time.RFC3339))
	kafkaServer, err := Cfg.String("kafka.server")
	if err != nil {
		fmt.Printf("%+v Kafka server not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaUsername, err := Cfg.String("kafka.credentials.username")
	if err != nil {
		fmt.Printf("%+v Kafka username not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaPassword, err := Cfg.String("kafka.credentials.password")
	if err != nil {
		fmt.Printf("%+v Kafka password not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaProtocol, err := Cfg.String("kafka.credentials.protocol")
	if err != nil {
		fmt.Printf("%+v Kafka protocol not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	kafkaMechanism, err := Cfg.String("kafka.credentials.mechanism")
	if err != nil {
		fmt.Printf("%+v Kafka authentication mechanism not defined.\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
		"sasl.username":     kafkaUsername,
		"sasl.password":     kafkaPassword,
		"security.protocol": kafkaProtocol,
		"sasl.mechanism":    kafkaMechanism,
	})
	if err != nil {
		fmt.Printf("%+v Kafka connection failed\n", time.Now().UTC().Format(time.RFC3339))
		panic(err)
	}
	fmt.Printf("%+v Kafka connected\n", time.Now().UTC().Format(time.RFC3339))
	return producer
}
