package main

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

const (
	natsURL = "127.0.0.1:4222"
	subject = "event.product"
)

func main() {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	request := []byte(`{"name":"Veermal", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request)
	if err != nil {
		logrus.Error("failed to publish request")
	}

	request2 := []byte(`{"name":"Bombite", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request2)
	if err != nil {
		logrus.Error("failed to publish request")
	}

	request3 := []byte(`{"name":"Elundo", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request3)
	if err != nil {
		logrus.Error("failed to publish request")
	}

	err = nc.Flush()
	if err != nil {
		logrus.Error("failed to flush nats")
	}
}
