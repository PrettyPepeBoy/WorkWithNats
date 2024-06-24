package main

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	nc, err := nats.Connect("127.0.0.1:4222")

	if err != nil {
		os.Exit(1)
	}
	request := []byte(`{"name":"Bughati", "category":"Car" ,"price":500000, "color":"dirty"}`)
	_, err = nc.Request("database.product.put", request, 10*time.Millisecond)
	request2 := []byte(`{"name":"Venom", "category":"Car" ,"price":500000, "color":"dirty"}`)
	_, err = nc.Request("database.product.put", request2, 10*time.Millisecond)
	request3 := []byte(`{"name":"Rafaelo", "category":"Car" ,"price":500000, "color":"dirty"}`)
	_, err = nc.Request("database.product.put", request3, 10*time.Millisecond)
	if err != nil {
		logrus.Error("failed to publish request")
	}
}
