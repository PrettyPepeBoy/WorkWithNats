package main

import (
	"TestTaskNats/internal/cache"
	"TestTaskNats/internal/config"
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/services/database/postgresservice"
	"TestTaskNats/internal/transport/endpoint"
	"TestTaskNats/internal/transport/natsserver/handlers/producthandler"
	"context"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.MustInitConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	strg := connectDatabase()
	natsConn := connectNats()
	cch := createNewCache()

	go doProductHandler(natsConn, strg)

	httpHandler := connectInternalServicesToHttpHandlers(cch, strg)

	logrus.Infof("listen server on port: %v", viper.GetString("http_server.port"))
	go func() {
		err := fasthttp.ListenAndServe(":"+viper.GetString("http_server.port"), httpHandler.CreateMux())
		if err != nil {
			logrus.Fatalf("failed to connect to http server")
		}
	}()
	<-ctx.Done()
	logrus.Info("stopping server")
}

func connectDatabase() *postgres.Storage {
	strg := postgres.MustConnectDB(context.Background(),
		viper.GetString("database.username"),
		os.Getenv(viper.GetString("database.password")),
		viper.GetString("database.host"),
		viper.GetString("database.database"),
		viper.GetInt("database.port"),
	)
	return strg
}

func connectNats() *nats.Conn {
	natsConn, err := nats.Connect(viper.GetString("natsserver.host") + ":" + viper.GetString("natsserver.port"))
	if err != nil {
		logrus.Fatal("failed to connect to nats")
	}
	return natsConn
}

func createNewCache() *cache.Cache {
	return cache.NewCache(viper.GetDuration("cache.cleanup_interval"))
}

func connectInternalServicesToHttpHandlers(cch *cache.Cache, strg *postgres.Storage) endpoint.HttpHandler {
	return endpoint.NewInternalServicesForHttpHandlers(cch, strg)
}

func doProductHandler(natsConn *nats.Conn, strg *postgres.Storage) {
	productHandler, err := producthandler.NewProductHandler(natsConn)
	if err != nil {
		logrus.Fatal("failed to connect to nats")
	}
	for {
		product, ok := <-productHandler.C
		if !ok {
			break
		}
		postgresservice.PutDataInTable(strg, product)
	}
}
