package main

import (
	"brightbox"
	"fmt"
	"os"
	"log"
	"io/ioutil"
	config "github.com/kless/goconfig/config"
)

var (
  ErrConfigUnreadable  = os.NewError("brightbox: couldn't read config file")
)

type Config struct {
	filename     string
	config       *config.Config
	auth_url     string
	client_id    string
	secret       string
	api_url      string
	auth         *brightbox.ApiClientAuth
	client       *brightbox.Client
}

func NewConfig() (*Config, os.Error) {
	var	err   os.Error
	conf := new(Config)
	conf.filename = os.Getenv("HOME") + "/.brightbox/config"
	conf.config, err = config.ReadDefault(conf.filename)
	if err != nil {
		return nil, ErrConfigUnreadable
	}
	client_name := os.Getenv("CLIENT")
	if client_name == "" {
		log.Fatal("You must specify the config section name in environment variable CLIENT")
	}
	conf.auth_url, _ = conf.config.String(client_name, "auth_url")
	conf.api_url, _ = conf.config.String(client_name, "api_url")
	conf.client_id, _ = conf.config.String(client_name, "client_id")
	conf.secret, _ = conf.config.String(client_name, "secret")
	if conf.auth_url == "" && conf.api_url != "" {
		conf.auth_url = conf.api_url
	}
	return conf, nil
}

// SetupAuthenticatorCache tries to read a cached token from the local filesystem
func (conf *Config) SetupAuthenticatorCache(auth brightbox.Authenticator) os.Error {
	var (
		err       os.Error
		token     string
		expires   int64
		f         *os.File
	)
	cache_filename := "/home/john/.brightbox/" + auth.String() + ".oauth_token.v2"
	f, err = os.Open(cache_filename)
	if f != nil {
		_, err = fmt.Fscanf(f, "%s", &token)
		_, err = fmt.Fscanf(f, "%d", &expires)
		f.Close()
	}
	if auth.SetToken(token, expires) != nil {
		token, expires, err = auth.Token()
		if err != nil {
			return nil
		}
		// BUG(johnl): should write to temp file and rename
		ioutil.WriteFile(cache_filename, []uint8(fmt.Sprintf("%s %d", token, expires)), 0600)
	}
	return nil
}

func (conf *Config) Auth() *brightbox.ApiClientAuth {
	if conf.auth == nil {
		conf.auth = brightbox.NewApiClientAuth(conf.auth_url, conf.client_id, conf.secret)
	}
	return conf.auth
}

func (conf *Config) Client() *brightbox.Client {
	if conf.client == nil {
		conf.client = brightbox.NewClient(conf.api_url, "1.0", conf.Auth())
	}
	return conf.client
}