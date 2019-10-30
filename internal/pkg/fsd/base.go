package fsd

import "strings"

// Base contains the common fields of all inter-server packets
type Base struct {
	Destination  string
	Source       string
	PacketNumber int
	HopCount     int
}

// reassemble puts the FSD packet back together for debugging
func reassemble(fields []string) string {
	return strings.Join(fields, ":")
}
