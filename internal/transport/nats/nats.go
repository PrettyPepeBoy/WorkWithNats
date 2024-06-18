package nats

import (
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/services/database/postgresservice"
	"TestTaskNats/internal/services/validation"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type NatsServer struct {
	Conn *nats.Conn
}

func MustConnectNats(strg *postgres.Storage) *NatsServer {
	err := viper.BindEnv("env", "NATS_URL")
	if err != nil {
		logrus.Fatalf("failed to get environment NATS_URL, error: %v", err)
	}

	nc, err := nats.Connect(viper.GetString("env"))
	if err != nil {
		logrus.Fatalf("failed to connect to nats server, error: %v", err)
	}
	natsServer := &NatsServer{Conn: nc}
	initSubscriptions(natsServer, strg)
	return natsServer
}

func initSubscriptions(nc *NatsServer, strg *postgres.Storage) {
	err := nc.addSubscription("database.put", func(m *nats.Msg) {
		if !validation.Valid(m.Data) {
			_ = m.Respond([]byte("failed validation"))
			return
		}
		postgresservice.PutDataInTable(strg, m.Data)
		_ = m.Respond([]byte("success"))
	})
	if err != nil {
		logrus.Fatalf("failed to add init subscription error: %v", err)

	}
}

func (nc *NatsServer) addSubscription(subject string, fn func(m *nats.Msg)) error {
	_, err := nc.Conn.Subscribe(subject, fn)
	if err != nil {
		return err
	}
	return nil
}
