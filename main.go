package main

import (
	"context"
	"errors"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/endpoint"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/product"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os/signal"
	"syscall"
)

var (
	productTable   *product.Table
	natsConn       *nats.Conn
	productsCache  *cache.Cache
	productHandler *product.Handler
)

func main() {
	var err error

	mustInitConfig()

	productsCache = cache.NewCache(viper.GetDuration("cache.cleanup_interval"))
	productTable, err = product.NewTable()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	connectNats()
	initProductProcessing()

	httpHandler := endpoint.NewHttpHandler(productsCache, productTable)

	logrus.Infof("listen server on port: %v", viper.GetString("http_server.port"))
	go func() {
		err := fasthttp.ListenAndServe(":"+viper.GetString("http-server.port"), httpHandler.Handle)
		if err != nil {
			logrus.Fatalf("failed to connect to http server %v", err.Error())
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
		logrus.Fatalf("failed to read config file, error: %v", err)
	}
}

func connectNats() {
	var err error
	natsConn, err = nats.Connect(viper.GetString("nats-server.host"))
	if err != nil {
		logrus.Fatalf("failed to connect to nats: %v", err)
	}

	natsConn.SetErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		if errors.Is(err, nats.ErrSlowConsumer) {
			logrus.Error(conn.ConnectedUrl(), " - ", subscription.Subject, " - ", err.Error())
		} else {
			logrus.Error("Unexpected nats error: ", err.Error())
		}
	})
}

func initProductProcessing() {
	var err error
	productHandler, err = product.NewHandler(natsConn)
	if err != nil {
		logrus.Fatal("failed to connect to nats")
	}

	go func() {
		for e := range productHandler.C {
			_, err := productTable.Put(e.Name, e.Data)
			if err != nil {
				logrus.Errorf("failed to put in table, error: %v", err.Error())
			}
		}
	}()
}
