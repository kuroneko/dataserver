package dataserver

import (
	"bufio"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"github.com/bugsnag/bugsnag-go"
	"net"
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
	clientList := dataserver.ClientList{}
	listen(bufReader, clientList, conn)
}

// listen continually reads, parses and handles FSD packets.
func listen(bufReader *bufio.Reader, clientList dataserver.ClientList, conn net.Conn) {
	for {
		bytes, err := fsd.ReadMessage(bufReader)
		if err != nil {
			_ = bugsnag.Notify(err)
			continue
		}
		split := fsd.ParseMessage(bytes)
		if split[0] == "ADDCLIENT" {
			err = dataserver.AddClient(split, &clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "RMCLIENT" {
			err = dataserver.RemoveClient(split, &clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "PD" {
			err = dataserver.UpdatePosition(split, &clientList)
			if err != nil {
				_ = bugsnag.Notify(err)
				continue
			}
		}
		if split[0] == "AD" {
			err = dataserver.UpdateControllerData(split, &clientList)
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
