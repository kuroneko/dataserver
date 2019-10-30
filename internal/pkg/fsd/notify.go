package fsd

import (
	"strconv"
	"strings"
)

// Notify NOTIFY
type Notify struct {
	Base
	FeedFlag int
	Ident    string
	Name     string
	Email    string
	Hostname string
	Version  string
	Flags    int
	Location string
}

// Serialize converts a struct into an FSD packet
func (n Notify) Serialize() string {
	msg := strings.Builder{}
	msg.WriteString("NOTIFY")
	msg.WriteString(":")
	msg.WriteString(n.Destination)
	msg.WriteString(":")
	msg.WriteString(n.Source)
	msg.WriteString(":")
	msg.WriteString("B")
	msg.WriteString(strconv.Itoa(n.PacketNumber))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(n.HopCount))
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(n.FeedFlag))
	msg.WriteString(":")
	msg.WriteString(n.Ident)
	msg.WriteString(":")
	msg.WriteString(n.Name)
	msg.WriteString(":")
	msg.WriteString(n.Email)
	msg.WriteString(":")
	msg.WriteString(n.Hostname)
	msg.WriteString(":")
	msg.WriteString(n.Version)
	msg.WriteString(":")
	msg.WriteString(strconv.Itoa(n.Flags))
	msg.WriteString(":")
	msg.WriteString(n.Location)
	return msg.String()
}
