package dataserver

import (
	"bufio"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"fmt"
	"github.com/bugsnag/bugsnag-go"
	"github.com/pkg/errors"
	"net"
	"time"
)

// Start connects and begins parsing and saving data files.
func Start() {
	dataserver.ReadConfig()
	dataserver.ConfigureBugsnag()
	conn := fsd.Connect()
	defer func() {
		if err := conn.Close(); err != nil {
			_ = bugsnag.Notify(err)
		}
	}()
	bufReader := fsd.SetupReader(conn)
	fsd.Sync(conn)
	clientList := &dataserver.ClientList{}
	go update()
	listen(bufReader, clientList, conn)
}

// update handles the creation of a one minute ticker for updating the data file
func update() {
	now := time.Now().UTC()
	for clientList := range dataserver.Channel {
		if time.Since(now) >= time.Minute {
			err := updateFile(clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
			}
			now = time.Now().UTC()
		}
	}
}

// listen continually reads, parses and handles FSD packets.
func listen(bufReader *bufio.Reader, clientList *dataserver.ClientList, conn net.Conn) {
	for {
		bytes, err := fsd.ReadMessage(bufReader)
		if err != nil {
			_ = bugsnag.Notify(err)
			continue
		}
		split := fsd.ParseMessage(bytes)
		if split[0] == "ADDCLIENT" {
			err = dataserver.AddClient(split, clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "RMCLIENT" {
			err = dataserver.RemoveClient(split, clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "PD" {
			err = dataserver.UpdatePosition(split, clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "AD" {
			err = dataserver.UpdateControllerData(split, clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "PLAN" {
			err = dataserver.UpdateFlightPlan(split, clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "PING" {
			fsd.Pong(conn, split)
		}
	}
}

// updateFile encodes the current clientList and prints to the data file
func updateFile(clientList dataserver.ClientList) error {
	clientJSON, err := dataserver.EncodeJSON(clientList)
	if err != nil {
		return errors.Wrapf(err, "Failed to encode client list to JSON %+v", clientList)
	}
	err = dataserver.WriteDataFile(clientJSON)
	if err != nil {
		return errors.Wrapf(err, "Failed to write JSON to file %+v", clientJSON)
	}
	fmt.Printf("%+v Data file updated\n", time.Now().UTC().Format(time.RFC3339))
	return nil
}
