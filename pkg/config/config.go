package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Messaging nsqConfig
	Storage s3Config
}

type s3Config struct {
	Host string
	Port int
	AccessKey string
	SecretKey string
	TLS bool
}

type nsqConfig struct {
	Host string
	Port string
	Topic string
	Concurrency int
	MaxAttempts int
	MaxInFlight int
	DefaultRequeueDelay string
}

var App Config

func (config *Config) Init() error {
	viper := viper.New()
	viper.SetConfigName("config")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	config.Messaging.Host = viper.GetString("Messaging.Host")
	config.Messaging.Port = viper.GetString("Messaging.Port")
	config.Messaging.Topic = viper.GetString("Messaging.Topic")
	config.Messaging.Concurrency = viper.GetInt("Messaging.Concurrency")
	config.Messaging.MaxAttempts = viper.GetInt("Messaging.MaxAttempts")
	config.Messaging.MaxInFlight = viper.GetInt("Messaging.MaxInFlight")
	config.Messaging.DefaultRequeueDelay = viper.GetString("Messaging.DefaultRequeueDelay")

	config.Storage.Host = viper.GetString("Storage.Host")
	config.Storage.Port = viper.GetInt("Storage.Port")
	config.Storage.AccessKey = viper.GetString("Storage.AccessKey")
	config.Storage.SecretKey = viper.GetString("Storage.SecretKey")
	config.Storage.TLS = viper.GetBool("Storage.TLS")

	return nil
}