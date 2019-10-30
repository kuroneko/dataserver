package fsd

import (
	"fmt"
	"reflect"
	"strconv"
	"syreclabs.com/go/faker"
	"testing"
)

func TestDeserializeAddClient(t *testing.T) {
	packetNumber := faker.Number().NumberInt(8)
	hopCount := faker.Number().NumberInt(1)
	cid := faker.Number().NumberInt(7)
	server := faker.Internet().DomainWord()
	callsign := faker.Internet().UserName()
	rating := faker.Number().NumberInt(1)
	name := faker.Name().Name()
	type args struct {
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		want    AddClient
		wantErr bool
	}{
		{"ATC", args{fields: []string{"ADDCLIENT", "*", server, fmt.Sprintf("B%v", packetNumber), strconv.Itoa(hopCount), strconv.Itoa(cid), server, callsign, "2", strconv.Itoa(rating), "100", name, "-1", "0"}}, AddClient{
			Base: Base{
				Destination:  "*",
				Source:       server,
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			CID:              cid,
			Server:           server,
			Callsign:         callsign,
			Type:             2,
			Rating:           rating,
			ProtocolRevision: 100,
			RealName:         name,
			SimType:          -1,
			Hidden:           0,
		}, false},
		{"Pilot", args{fields: []string{"ADDCLIENT", "*", server, fmt.Sprintf("B%v", packetNumber), strconv.Itoa(hopCount), strconv.Itoa(cid), server, callsign, "1", "1", "100", name, "-1", "0"}}, AddClient{
			Base: Base{
				Destination:  "*",
				Source:       server,
				PacketNumber: packetNumber,
				HopCount:     hopCount,
			},
			CID:              cid,
			Server:           server,
			Callsign:         callsign,
			Type:             1,
			Rating:           1,
			ProtocolRevision: 100,
			RealName:         name,
			SimType:          -1,
			Hidden:           0,
		}, false},
		{"Not enough fields", args{fields: []string{"*", server, fmt.Sprintf("B%v", packetNumber), strconv.Itoa(hopCount), strconv.Itoa(cid), server, callsign, "1", "1", "100", name, "-1", "0"}}, AddClient{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeserializeAddClient(tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeAddClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeAddClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}
