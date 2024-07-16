package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/cache"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/endpoint"
	"github.com/PrettyPepeBoy/WorkWithNats/internal/objects/product"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	natsConn       *nats.Conn
	productCache   *cache.Cache[cache.Int, cache.ByteSlc]
	productTable   *product.Table
	productHandler *product.Handler
)

func main() {
	var err error

	mustInitConfig()
	mustConnectNats()

	productCache = cache.NewCache[cache.Int, cache.ByteSlc]()
	productTable, err = product.NewTable()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	httpHandler := endpoint.NewHttpHandler(productCache, productTable)
	initProductProcessing()

	logrus.Infof("listen server on port: %v", viper.GetString("http-server.port"))
	go func() {
		err := fasthttp.ListenAndServe(":"+viper.GetString("http-server.port"), httpHandler.Handle)
		if err != nil {
			logrus.Fatalf("failed to connect to http server")
		}
	}()

	initBackupCache()

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

func initBackupCache() {
	t := viper.GetDuration("cache.backup_interval")

	go func() {
		for {
			time.Sleep(t)

			filename := fmt.Sprintf("/var/lib/cache/data/cache_data_%v.gz", time.Now().UnixNano())

			file, err := os.Create(filename)
			if err != nil {
				logrus.Errorf("failed to create gzip file for backup, error: %v", err)
				return
			}

			gzipWriter := gzip.NewWriter(file)
			bufWriter := bufio.NewWriter(gzipWriter)
			productCache.GetAllRawData(bufWriter)

			err = bufWriter.Flush()
			if err != nil {
				logrus.Errorf("failed to flush buffer, error: %v", err)
				_ = file.Close()
				_ = os.Remove(filename)
				continue
			}

			err = gzipWriter.Close()
			if err != nil {
				logrus.Errorf("failed to close gzipWriter, error: %v", err)
				_ = file.Close()
				_ = os.Remove(filename)
				continue
			}

			_ = file.Close()
		}
	}()
}
