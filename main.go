package main

import (
	"TestTaskNats/internal/cache"
	"TestTaskNats/internal/config"
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/transport/endpoint"
	"TestTaskNats/internal/transport/nats"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.MustInitConfig()
	c := cache.InitCache(viper.GetDuration("cache.cleanup_interval"))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	strg := postgres.MustConnectDB(ctx,
		viper.GetString("database.username"),
		os.Getenv(viper.GetString("database.password")),
		viper.GetString("database.host"),
		viper.GetString("database.database"),
		viper.GetInt("database.port"),
	)
	ns := nats.MustConnectNats(strg)

	rawbyte := []byte(`{"name":"TOYOTA", "price":500000, "amount":70}`)
	_, err := ns.Conn.Request("database.put", rawbyte, 100*time.Millisecond)
	if err != nil {
		logrus.Errorf("failed to send request to nats server, error: %v", err)
	}

	logrus.Infof("listen server on port: %v", viper.GetString("http_server.port"))
	go func() {
		err := fasthttp.ListenAndServe(":"+viper.GetString("http_server.port"), endpoint.CreateMux(strg, c))
		if err != nil {
			logrus.Fatalf("failed to connect to http server")
		}
	}()
	<-ctx.Done()
	logrus.Info("stopping server")
}
