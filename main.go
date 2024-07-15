package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
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
	productCache   *cache.Cache[int, []byte]
	productTable   *product.Table
	productHandler *product.Handler
)

func main() {
	var err error

	mustInitConfig()
	mustConnectNats()

	productCache = cache.NewCache[int, []byte]()
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
	initBackupDatabase()

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
	bucketsAmount := viper.GetInt("cache.buckets_amount")

	go func() {
		for {
			time.Sleep(t)

			bytesBuf := make([]byte, 0, 2<<12)
			buf := bytes.NewBuffer(bytesBuf)
			for i := 0; i < bucketsAmount; i++ {
				rawBytes, err := productCache.Buckets[i].GetAllRawData()
				if err != nil {
					logrus.Errorf("failed to get raw data from cache, error: %v", err)
					return
				}
				buf.Write(rawBytes)
			}

			file, err := os.Create("/var/lib/cache/data/" + "cache_data.gz")
			if err != nil {
				logrus.Errorf("failed to create gzip file for backup, error: %v", err)
				return
			}

			w := gzip.NewWriter(file)
			_, err = w.Write(buf.Bytes())
			if err != nil {
				logrus.Errorf("failed to write data to gzip, error: %v", err)
				return
			}

			_ = w.Close()
		}
	}()
}

func initBackupDatabase() {
	t := viper.GetDuration("cache.backup_interval")

	go func() {
		for {
			time.Sleep(t)
			var buf bytes.Buffer
			encoder := gob.NewEncoder(&buf)

			rows, err := productTable.GetAllFromTable()
			if err != nil {
				logrus.Errorf("failed to get data from table, error: %v", err)
				return
			}

			var id uint32
			var data []byte
			for rows.Next() {
				err = rows.Scan(&id, &data)
				if err != nil {
					logrus.Errorf("failed to scan from row, error: %v", err)
					return
				}

				err = encoder.Encode(id)
				if err != nil {
					logrus.Errorf("failed to encode id, error: %v", err)
				}

				err = encoder.Encode(data)
				if err != nil {
					logrus.Errorf("failed to encode data, error: %v", err)
				}
			}

			file, err := os.Create("/var/lib/database/data/" + "database_data.gz")
			if err != nil {
				logrus.Errorf("failed to create gzip file for backup, error: %v", err)
				return
			}

			w := gzip.NewWriter(file)
			_, err = w.Write(buf.Bytes())
			if err != nil {
				logrus.Errorf("failed to write data to gzip, error: %v", err)
				return
			}

			_ = w.Close()
		}
	}()
}
