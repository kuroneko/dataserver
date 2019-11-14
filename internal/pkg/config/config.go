package config

import (
	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/sentry-go"
	"github.com/olebedev/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// Cfg contains all of the necessary configuration data.
type Config struct {
	FSD               *FsdConfig
	S3Buckets         []*BucketConfig
	Kafka             *KafkaConfig
	DataFileDirectory string
}

func New(configFile string) (c *Config, err error) {
	c = &Config{}
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	yamlString := string(file)
	configObj, err := config.ParseYaml(yamlString)
	if err != nil {
		return nil, err
	}
	// parse the items.

	c.DataFileDirectory, err = configObj.String("data.file.directory")
	if err != nil {
		return nil, err
	}
	err = configureSentry(configObj)
	if err != nil {
		return nil, err
	}
	c.FSD, err = newFsdConfig(configObj)
	if err != nil {
		return nil, err
	}
	c.S3Buckets, err = allBucketConfigs(configObj)
	if err != nil {
		return nil, err
	}
	c.Kafka, err = newKafkaConfigFromConfig(configObj)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ConfigureSentry sets up Sentry for panic reporting
func configureSentry(c *config.Config) (err error) {
	dsn, err := c.String("sentry.credentials.dsn")
	if err != nil {
		// if the DSN isn't set, that's not a fatal error.  Just don't enable Sentry.
		return nil
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		return err
	}
	hook, err := logrus_sentry.NewSentryHook(dsn, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})
	if err != nil {
		return err
	}
	log.AddHook(hook)
	return nil
}
