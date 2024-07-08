package main

import (
	"crypto/rand"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"math/big"
)

const (
	natsURL                = "127.0.0.1:4222"
	subject                = "event.product"
	defaultAmountOfNumbers = 3
	goroutinesAmount       = 10
)

var letters = "abcdefghijklmnopqwstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type DefaultRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
	Price    int    `json:"price"`
}

func main() {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	//iterations, err := rand.Int(rand.Reader, big.NewInt(600))
	//if err != nil {
	//	logrus.Fatal("failed to generate word")
	//}

	for i := 0; i < 600; i++ {
		rawByte := prepareDefaultRequest()
		err = nc.Publish(subject, rawByte)
		if err != nil {
			logrus.Error("failed to publish request")
		}
	}

	err = nc.Flush()
	if err != nil {
		logrus.Error("failed to flush nats")
	}

}

func prepareDefaultRequest() []byte {
	slc := make([]string, defaultAmountOfNumbers)
	var req DefaultRequest
	for i := 0; i < defaultAmountOfNumbers; i++ {
		slc[i] = generateRandomWord()
	}

	req.Name = slc[0]
	req.Color = slc[1]
	req.Category = slc[2]
	nBig, err := rand.Int(rand.Reader, big.NewInt(1<<35))
	if err != nil {
		logrus.Fatal("failed to generate word")
	}
	req.Price = int(nBig.Int64())

	rawByte, err := json.Marshal(req)
	if err != nil {
		logrus.Fatal("failed to marshal json, error: ", err)
	}
	return rawByte
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
