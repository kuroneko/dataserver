package dataserver

import (
	"bufio"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"net"
	"time"
)

// Start connects and begins parsing and saving data files.
func Start() {
	dataserver.ReadConfig()
	dataserver.ConfigureSentry()
	conn := fsd.Connect()
	defer func() {
		if err := conn.Close(); err != nil {
			sentry.CaptureException(err)
		}
	}()
	bufReader := fsd.SetupReader(conn)
	fsd.Sync(conn)
	clientList := &dataserver.ClientList{}
	go update()
	listen(bufReader, clientList, conn)
}

// update handles the creation of a 15 second ticker for updating the data file
func update() {
	now := time.Now().UTC()
	for clientList := range dataserver.Channel {
		if time.Since(now) >= (15 * time.Second) {
			err := updateFile(clientList)
			if err != nil {
				sentry.CaptureException(err)
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
			sentry.CaptureException(err)
			continue
		}
		split := fsd.ParseMessage(bytes)
		if split[0] == "ADDCLIENT" && len(split) >= 12 {
			err = dataserver.AddClient(split, clientList)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}
		}
		if split[0] == "RMCLIENT" && len(split) >= 6 {
			err = dataserver.RemoveClient(split, clientList)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}
		}
		if split[0] == "PD" && len(split) >= 13 {
			err = dataserver.UpdatePosition(split, clientList)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}
		}
		if split[0] == "AD" && len(split) >= 12 {
			err = dataserver.UpdateControllerData(split, clientList)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}
		}
		if split[0] == "PLAN" && len(split) >= 22 {
			err = dataserver.UpdateFlightPlan(split, clientList)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}
		}
		if split[0] == "PING" && len(split) >= 6 {
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
