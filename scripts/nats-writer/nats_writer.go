package main

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

const (
	natsUrl = "127.0.0.1:4222"
	subject = "events.products"
)


func main() {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	request := []byte(`{"name":"Bughati", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	request = []byte(`{"name":"Venom", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	request = []byte(`{"name":"Rafaelo", "category":"Car" ,"price":500000, "color":"dirty"}`)
	err = nc.Publish(subject, request)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = nc.Flush()
	if err != nil {
		logrus.Fatal(err.Error())
	}
}
