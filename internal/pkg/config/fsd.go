package config

import (
	"fmt"
	"github.com/olebedev/config"
)

type FsdConfig struct {
	Hostname string
	Port     uint16

	DataServerName     string
	DataServerEmail    string
	DataServerLocation string
}

func (fc *FsdConfig) String() string {
	return fmt.Sprintf("%v:%v", fc.Hostname, fc.Port)
}

func newFsdConfig(c *config.Config) (fc *FsdConfig, err error) {
	fc = &FsdConfig{}
	fc.Hostname, err = c.String("fsd.server.ip")
	if err != nil {
		return nil, err
	}
	fc.Port = uint16(c.UInt("fsd.server.port", 4113))
	fc.DataServerName = c.UString("data.server.name", "DSERVERNG")

	fc.DataServerEmail, err = c.String("data.server.email")
	if err != nil {
		return nil, err
	}
	fc.DataServerLocation, err = c.String("data.server.location")
	if err != nil {
		return nil, err
	}
	return
}
