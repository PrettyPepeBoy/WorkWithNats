package main

import (
	"crypto/rand"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

const (
	natsURL = "127.0.0.1:4222"
	subject = "event.product"
	limit   = 1000
)

var colors = []string{"red", "yellow", "green", "blue", "brown", "black", "white", "pink", "magenta", "purple"}
var category = []string{"car"}
var location = []string{"Yekaterinburg", "Moscow", "Saint-Petersburg", "Kazan", "Novosibirsk"}

var letters = "abcdefghijklmnopqwstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type DefaultRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
	Location string `json:"location"`
	Price    int    `json:"price"`
	Amount   int    `json:"amount"`
}

func main() {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	var count int
	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			if count == limit {
				return
			}

			rawByte := prepareDefaultRequest()
			err = nc.Publish(subject, rawByte)
			if err != nil {
				logrus.Error("failed to publish request")
			}
			count++
		}
	}()

	<-ticker.C
	err = nc.Flush()
	if err != nil {
		logrus.Error("failed to flush nats")
	}

	logrus.Infof("final count: %v", count)
	logrus.Info("finish ticker")
}

func prepareDefaultRequest() []byte {
	var req DefaultRequest

	req.Name = generateRandomWord()
	req.Color = colors[generateRandomNumber(len(colors))]
	req.Category = category[0]
	req.Location = location[generateRandomNumber(len(location))]

	nBig, err := rand.Int(rand.Reader, big.NewInt(1<<20))
	if err != nil {
		logrus.Fatal("failed to generate number")
	}
	req.Price = int(nBig.Int64())

	nBig, err = rand.Int(rand.Reader, big.NewInt(1<<20))
	if err != nil {
		logrus.Fatal("failed to generate number")
	}
	req.Amount = int(nBig.Int64())

	rawByte, err := json.Marshal(req)
	if err != nil {
		logrus.Fatal("failed to marshal json, error: ", err)
	}
	return rawByte
}

func generateRandomNumber(max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-1)))
	if err != nil {
		logrus.Fatal("failed to generate word")
	}

	return int(nBig.Int64())
}

func generateRandomWord() string {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(24)))
	if err != nil {
		logrus.Fatal("failed to generate word")
	}

	var word string

	for i := 0; i < int(nBig.Int64()); i++ {
		nm, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			logrus.Fatal("failed to generate word")
		}
		word += string(letters[int(nm.Int64())])
	}

	return word
}
