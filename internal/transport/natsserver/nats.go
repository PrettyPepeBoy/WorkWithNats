package natsserver

import (
	"TestTaskNats/internal/models"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
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

	h.natsSubs, _ = natsConn.Subscribe("database.product.put", h.Process)
	return h, nil
}

func (h *ProductHandler) Process(msg *nats.Msg) {
	_ = msg.Respond([]byte("get message"))
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
	logrus.Info(h.C, "chanel for reading")
	logrus.Info(h.innerChannel, "inner chanel")
}

func validateProductData(product models.ProductBody) bool {
	if !correctSymbols(product.Name, product.Category) {
		return false
	}
	return true
}

func correctSymbols(str ...string) bool {
	for _, word := range str {
		for _, elem := range word {
			if !((elem >= 'a' && elem <= 'z') || (elem >= 'A' && elem <= 'Z')) {
				return false
			}
		}
	}
	return true
}
