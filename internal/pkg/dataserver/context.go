package dataserver

import (
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net/textproto"
)

// Context holds the application current context
type Context struct {
	Consumer   *textproto.Conn
	Producer   *kafka.Producer
	ClientList *ClientList
}
