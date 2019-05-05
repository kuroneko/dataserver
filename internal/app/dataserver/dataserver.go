package dataserver

import (
	"bufio"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"net"
)

// Start connects and begins parsing and saving data files.
func Start() {
	dataserver.ReadConfig()
	conn := fsd.Connect()
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
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
		bytes := fsd.ReadMessage(bufReader)
		split := fsd.ParseMessage(bytes)
		if split[0] == "ADDCLIENT" {
			dataserver.AddClient(split, &clientList)
		}
		if split[0] == "RMCLIENT" {
			dataserver.RemoveClient(split, &clientList)
		}
		if split[0] == "PD" {
			dataserver.UpdatePosition(split, &clientList)
		}
		if split[0] == "AD" {
			dataserver.UpdateControllerData(split, &clientList)
		}
		if split[0] == "PING" {
			fsd.Pong(conn, split)
		}
	}
}
