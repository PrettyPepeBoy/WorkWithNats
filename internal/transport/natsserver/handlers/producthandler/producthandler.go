package producthandler

import (
	"TestTaskNats/internal/models"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ProductHandler struct {
	C            <-chan models.ProductBody
	innerChannel chan models.ProductBody

	natsSubs *nats.Subscription
}

func NewProductHandler(natsConn *nats.Conn) (*ProductHandler, error) {
	c := make(chan models.ProductBody)

	h := &ProductHandler{
		C:            c,
		innerChannel: c,
	}

	var err error
	subject := viper.GetString("nats_server.subject")
	h.natsSubs, err = natsConn.Subscribe(subject, h.Process)
	if err != nil {
		logrus.Errorf("[NewProductHandler] failed to subscribe to %s, error: %v", subject, err)
	}

	return h, nil
}

func (h *ProductHandler) Process(msg *nats.Msg) {
	var product models.ProductBody
	err := json.Unmarshal(msg.Data, &product)
	if err != nil {
		logrus.Error("failed to unmarshal message data to product, error", err)
		return
	}
	if !validateProductData(product) {
		logrus.Error("failed to validate product", err)
		return
	}
	h.innerChannel <- product
}

func validateProductData(product models.ProductBody) bool {
	if !correctSymbols(product.Name, product.Category, product.Location, product.Color) {
		return false
	}
	if !correctInteger(product.Amount, product.Price) {
		return false
	}
	return true
}

func correctSymbols(str ...string) bool {
	for _, word := range str {
		if len(word) == 0 {
			continue
		}
		for _, elem := range word {
			if !((elem >= 'a' && elem <= 'z') || (elem >= 'A' && elem <= 'Z')) {
				return false
			}
		}
	}
	return true
}

func correctInteger(integers ...int) bool {
	for _, i := range integers {
		if i < 0 {
			return false
		}

		if i > 1<<35 {
			return false
		}
	}

	return true
}
