package cli

import (
	"../brightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type Config struct {
	Conn *brightbox.Connection
	App  *kingpin.Application
}

func (c *Config) Configure() error {
	c.Conn = &brightbox.Connection{
		Token:        os.Getenv("BRIGHTBOX_TOKEN"),
		AccountId:    os.Getenv("BRIGHTBOX_ACCOUNT"),
		ClientID:     os.Getenv("BRIGHTBOX_CLIENT_ID"),
		ClientSecret: os.Getenv("BRIGHTBOX_CLIENT_SECRET"),
		ApiUrl:       os.Getenv("BRIGHTBOX_API_URL"),
	}
	return nil
}

func NewConfig() (*Config, error) {
	var config Config
	err := config.Configure()
	return &config, err
}

func NewConfigAndConnect() (*Config, error) {
	config, err := NewConfig()
	if err != nil {
		return config, err
	}
	err = config.Conn.Connect()
	if err != nil {
		return config, err
	}
	return config, err
}
