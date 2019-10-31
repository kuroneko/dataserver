package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
)

// HandleATISData updates controller information
func (c *Context) HandleATISData(fields []string) error {
	atis, err := fsd.DeserializeATISData(fields)
	if err != nil {
		return err
	}
	c.ClientList.Mutex.Lock()
	defer c.ClientList.Mutex.Unlock()
	for i, v := range c.ClientList.ATCData {
		if v.Callsign == atis.From {
			switch atis.Type {
			case "T":
				c.handleATISText(v, i, atis)
				break
			case "E":
				*&c.ClientList.ATCData[i].ATISReceived = true
			}
			kafkaPush(c.Producer, c.ClientList.ATCData[i], "update_controller_data")
		}
	}
	log.WithFields(log.Fields{
		"callsign": atis.From,
		"data":     atis.Data,
	}).Debug("ATIS data packet received.")
	Channel <- *c.ClientList
	return nil
}

// handleATISText updates a controller's information
func (c *Context) handleATISText(v ATC, i int, atis fsd.ATISData) {
	switch v.ATISReceived {
	case true:
		*&c.ClientList.ATCData[i].ATIS = atis.Data
		*&c.ClientList.ATCData[i].ATISReceived = false
		break
	case false:
		*&c.ClientList.ATCData[i].ATIS += "\n" + atis.Data
		break
	}
}
