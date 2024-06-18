package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func MustInitConfig() {
	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("failed to read config file, error: %v", err)
	}
}
