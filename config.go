package main

import (
	"brightbox"
	"fmt"
	"os"
	"io/ioutil"
	config "github.com/kless/goconfig/config"
)

var (
  ErrConfigUnreadable         = os.NewError("brightbox: couldn't read config file")
)

type Config struct {
	path         string
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

	err = conf.EnsurePath()
	if err != nil {
		return nil, err
	}

	conf.filename = conf.path + "/" + "config"
	conf.config, err = config.ReadDefault(conf.filename)

	if err != nil {
		return nil, ErrConfigUnreadable
	}
	client_name := os.Getenv("CLIENT")
	if client_name == "" {
		if len(conf.config.Sections()) == 2 {	// built in "default" section + one configured api client
			client_name = conf.config.Sections()[1]
		} else if len(conf.config.Sections()) > 2 {
			client_name, err = conf.config.String("core", "default_client")
		}
	}

	if client_name == "" {
		return nil, os.NewError("Couldn't find an api client")
	}
	conf.api_url, err = conf.config.String(client_name, "api_url")
	if err != nil {	return nil, err	}

	conf.auth_url, err = conf.config.String(client_name, "auth_url")
	if conf.auth_url == "" && conf.api_url != "" {
		conf.auth_url = conf.api_url
	}

	conf.client_id, err = conf.config.String(client_name, "client_id")
	if err != nil {	return nil, err	}

	conf.secret, err = conf.config.String(client_name, "secret")
	if err != nil {	return nil, err	}

	return conf, nil
}

func (conf *Config) EnsurePath() os.Error {
	var err os.Error

	if conf.path != "" {
		return nil
	}
	path := os.Getenv("HOME") + "/.brightbox"

	stat, err := os.Stat(path)
	if err != nil {
		return os.NewError("Cannot create config path ("+path+"): " + err.String())
	}

	if stat.IsDirectory() == false {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return os.NewError("Cannot create config path ("+path+"): " + err.String())
		}
	}

	conf.path = path
	return nil
}

// SetupAuthenticatorCache tries to read a cached token from the local filesystem
func (conf *Config) SetupAuthenticatorCache(auth brightbox.Authenticator) os.Error {
	var (
		err       os.Error
		token     string
		expires   int64
		f         *os.File
	)
	conf.EnsurePath()
	cache_filename := conf.path + "/" + auth.String() + ".oauth_token.v2"
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