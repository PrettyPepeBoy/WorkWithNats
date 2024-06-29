package product

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Handler struct {
	C            <-chan Event
	innerChannel chan Event

	natsSubs *nats.Subscription
}

type Product struct {
	Id       uint32 `json:"id"`
	Name     string `json:"name,required"`
	Category string `json:"category,required"`
	Location string `json:"location,omitempty"`
	Color    string `json:"color,omitempty"`
	Price    uint32 `json:"price,omitempty"`
	Amount   uint32 `json:"amount,omitempty"`
}

type Event struct {
	Name string
	Data []byte
}

func NewHandler(natsConn *nats.Conn) (*Handler, error) {
	c := make(chan Event)

	h := &Handler{
		C:            c,
		innerChannel: c,
	}

	var err error
	subject := viper.GetString("nats-server.subjects.product")
	h.natsSubs, err = natsConn.Subscribe(subject, h.Process)
	if err != nil {
		logrus.Errorf("[NewHandler] failed to subscribe to %s, error: %v", subject, err)
	}

	return h, nil
}

func (h *Handler) Process(msg *nats.Msg) {
	var product Product
	logrus.Infof("recieved message: %s", string(msg.Data))
	err := json.Unmarshal(msg.Data, &product)
	if err != nil {
		logrus.Warn("failed to unmarshal message data to product, error: ", err)
		return
	}

	if !h.validateProductData(product) {
		logrus.Warn("failed to validate product, error: ", err)
		return
	}

	h.innerChannel <- Event{
		Name: product.Name,
		Data: msg.Data,
	}
}

func (h *Handler) validateProductData(product Product) bool {
	if len(product.Name) == 0 {
		return false
	}

	if len(product.Category) == 0 {
		return false
	}

	if !h.correctSymbols(product.Name) {
		return false
	}
	if !h.correctSymbols(product.Location) {
		return false
	}
	if !h.correctSymbols(product.Color) {
		return false
	}
	if !h.correctSymbols(product.Category) {
		return false
	}
	return true
}

func (h *Handler) correctSymbols(word string) bool {
	for _, elem := range word {
		if !((elem >= 'a' && elem <= 'z') || (elem >= 'A' && elem <= 'Z')) {
			return false
		}
	}
	return true
}
