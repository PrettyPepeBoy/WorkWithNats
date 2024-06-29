package main

import (
	"context"
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/endpoint"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os/signal"
	"syscall"
)

var (
	natsConn       *nats.Conn
	productCache   *cache.Cache[int, []byte]
	productTable   *product.Table
	productHandler *product.Handler
)

func main() {
	var err error

	mustInitConfig()
	mustConnectNats()

	productCache = cache.NewCache[int, []byte](viper.GetInt("cache.cleanup"))
	productTable, err = product.NewTable()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	httpHandler := endpoint.NewHttpHandler(productCache, productTable)
	initProductProcessing()
	logrus.Info("our cache cleanup_interval is: ", viper.GetDuration("cache.cleanup_interval"))
	logrus.Infof("listen server on port: %v", viper.GetString("http-server.port"))
	go func() {
		err := fasthttp.ListenAndServe(":"+viper.GetString("http-server.port"), httpHandler.Handle)
		if err != nil {
			logrus.Fatalf("failed to connect to http server")
		}
	}()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()

	logrus.Info("stopping server")
}

func mustInitConfig() {
	viper.SetConfigFile("./configuration.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("failed to read in config from configuration.yaml, error: %v", err)
	}
}

func mustConnectNats() {
	var err error
	natsConn, err = nats.Connect(viper.GetString("nats-server.host"))
	if err != nil {
		logrus.Fatalf("failed to connect to nats server, error: %v", err)
	}

	natsConn.SetErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		if errors.Is(err, nats.ErrSlowConsumer) {
			logrus.Error(conn.ConnectedUrl(), " - ", subscription.Subject, " - ", err.Error())
		} else {
			logrus.Error("unexpected nats error: ", err.Error())
		}
	})
}

func initProductProcessing() {
	var err error
	productHandler, err = product.NewHandler(natsConn)
	if err != nil {
		logrus.Fatalf("failed to connect to nats, error: %v", err)
	}

	go func() {
		for event := range productHandler.C {
			_, err = productTable.Put(event.Name, event.Data)
			if err != nil {
				logrus.Errorf("failed to put in table, error: %v", err)
			}
		}
	}()
}
