package config

import (
	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/sentry-go"
	"github.com/olebedev/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
)

// Cfg contains all of the necessary configuration data.
var Cfg *config.Config

// ConfigureKafka sets the Kafka connection up
func ConfigureKafka() *kafka.Producer {
	log.Info("Starting Kafka connection.")
	kafkaServer, err := Cfg.String("kafka.server")
	if err != nil {
		log.Fatal("Kafka server not defined.")
	}
	kafkaUsername, err := Cfg.String("kafka.credentials.username")
	if err != nil {
		log.Fatal("Kafka username not defined.")
	}
	kafkaPassword, err := Cfg.String("kafka.credentials.password")
	if err != nil {
		log.Fatal("Kafka password not defined.")
	}
	kafkaProtocol, err := Cfg.String("kafka.credentials.protocol")
	if err != nil {
		log.Fatal("Kafka protocol not defined.")
	}
	kafkaMechanism, err := Cfg.String("kafka.credentials.mechanism")
	if err != nil {
		log.Fatal("Kafka authentication mechanism not defined.")
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   kafkaServer,
		"sasl.username":       kafkaUsername,
		"sasl.password":       kafkaPassword,
		"security.protocol":   kafkaProtocol,
		"sasl.mechanism":      kafkaMechanism,
		"go.delivery.reports": false,
	})
	if err != nil {
		log.Fatal("Failed to connect to Kafka.")
	}
	log.Info("Kafka successfully connected.")
	return producer
}

// ConfigureSentry sets up Sentry for panic reporting
func ConfigureSentry() {
	dsn, err := Cfg.String("sentry.credentials.dsn")
	if err != nil {
		log.Fatal("Sentry DSN not defined.")
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		log.Fatal("Failed to initialize Sentry.")
	}
	hook, err := logrus_sentry.NewSentryHook(dsn, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})
	if err != nil {
		log.Fatal("Failed to setup Sentry log hook.")
	}
	log.AddHook(hook)
}

// ReadConfig reads the config file and instantiates the config object
func ReadConfig() {
	file, err := ioutil.ReadFile("configs/config.yml")
	if err != nil {
		log.Fatal("Failed to open configuration file.")
	}
	yamlString := string(file)
	Cfg, err = config.ParseYaml(yamlString)
	if err != nil {
		log.Fatal("Failed to parse configuration file.")
	}
}
