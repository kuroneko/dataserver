package fsd

import (
	"strconv"
	"strings"
)

// Pong PONG
type Pong struct {
	Base
	Data string
}

// Serialize converts a struct into an FSD packet
func (p Pong) Serialize() string {
	msg := strings.Builder{}
	msg.WriteString("PONG")
	msg.WriteString(":")
	msg.WriteString(p.Destination)
	msg.WriteString(":")
	msg.WriteString(p.Source)
	msg.WriteString(":")
	msg.WriteString("B")
	msg.WriteString(strconv.Itoa(p.PacketNumber))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(p.HopCount))
	msg.WriteString(":")
	msg.WriteString(p.Data)
	return msg.String()
}
