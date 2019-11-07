package dataserver

import (
	"dataserver/internal/pkg/fsd"
	log "github.com/sirupsen/logrus"
	"time"
)

// ATC is data about individual controllers on the network.
type ATC struct {
	Server       string     `json:"server"`
	Callsign     string     `json:"callsign"`
	Member       MemberData `json:"member"`
	Rating       int        `json:"rating"`
	Frequency    int        `json:"frequency"`
	FacilityType int        `json:"facility"`
	VisualRange  int        `json:"range"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	ATIS         string     `json:"atis"`
	ATISReceived bool       `json:"-"`
	LastUpdated  time.Time  `json:"last_updated"`
}

// HandleATCData updates a controllers's data in the Client list and updates the JSON file.
func (c *Context) HandleATCData(fields []string) error {
	atcData, err := fsd.DeserializeATCData(fields)
	if err != nil {
		return err
	}
	c.ClientList.Mutex.Lock()
	defer c.ClientList.Mutex.Unlock()
	for i, v := range c.ClientList.ATCData {
		if v.Callsign == atcData.Callsign {
			timeBetweenATCUpdates.Observe(time.Since(*&c.ClientList.ATCData[i].LastUpdated).Seconds())
			*&c.ClientList.ATCData[i].Frequency = atcData.Frequency
			*&c.ClientList.ATCData[i].FacilityType = atcData.FacilityType
			*&c.ClientList.ATCData[i].VisualRange = atcData.VisualRange
			*&c.ClientList.ATCData[i].Latitude = atcData.Latitude
			*&c.ClientList.ATCData[i].Longitude = atcData.Longitude
			*&c.ClientList.ATCData[i].LastUpdated = time.Now().UTC()
			kafkaPush(c.Producer, c.ClientList.ATCData[i], "update_controller_data")
			break
		}
	}
	log.WithFields(log.Fields{
		"callsign":  atcData.Callsign,
		"latitude":  atcData.Latitude,
		"longitude": atcData.Longitude,
	}).Debug("ATC data packet received.")
	Channel <- *c.ClientList
	return nil
}

// sendATCData sends fake ATC data for our client
func (c *Context) sendATCData(name string) {
	atcData := fsd.ATCData{
		Base: fsd.Base{
			Destination:  "*",
			Source:       name,
			PacketNumber: fsd.PdCount,
			HopCount:     1,
		},
		Callsign:     name,
		Frequency:    99999,
		FacilityType: 1,
		VisualRange:  100,
		Rating:       1,
		Latitude:     0.00000,
		Longitude:    0.00000,
	}
	err := fsd.Send(c.Consumer, atcData.Serialize())
	if err != nil {
		log.WithFields(log.Fields{
			"connection": c.Consumer,
			"error":      err,
		}).Fatal("Failed to send AD packet to FSD server.")
	}
	log.WithField("packet", atcData.Serialize()).Debug("Successfully sent AD packet to server.")
}
