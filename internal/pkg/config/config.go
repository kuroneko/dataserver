package config

import (
	"github.com/getsentry/sentry-go"
	"github.com/olebedev/config"
	"io/ioutil"
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
