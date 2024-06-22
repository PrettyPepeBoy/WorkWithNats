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
	request := []byte(`{"name":"Jigul", "category":"Car" ,"price":500000, "color":"dirty"}`)
	_, err = nc.Request("database.product.put", request, time.Second)
	if err != nil {
		logrus.Error("failed to send request")
	}
}
