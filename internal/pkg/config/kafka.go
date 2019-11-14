package config

import (
	"github.com/olebedev/config"
)

type KafkaConfig struct {
	ServerName       string
	SecurityProtocol string
	SaslUsername     string
	SaslPassword     string
	SaslMechanism    string
}

func newKafkaConfigFromConfig(cfg *config.Config) (kc *KafkaConfig, err error) {
	kc = &KafkaConfig{}
	kc.ServerName, err = cfg.String("kafka.server")
	if err != nil {
		return nil, err
	}
	kc.SecurityProtocol, err = cfg.String("kafka.credentials.mechanism")
	if err != nil {
		return nil, err
	}
	kc.SaslUsername, err = cfg.String("kafka.credentials.username")
	if err != nil {
		return nil, err
	}
	kc.SaslPassword, err = cfg.String("kafka.credentials.password")
	if err != nil {
		return nil, err
	}
	kc.SaslMechanism, err = cfg.String("kafka.credentials.mechanism")
	if err != nil {
		return nil, err
	}
	return kc, nil
}
