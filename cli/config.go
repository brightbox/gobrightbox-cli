package cli

import (
	"../brightbox"
	"fmt"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
	"os/user"
	"path"
	"sort"
)

type ConfigClient struct {
	ClientName     string
	ClientID       string `ini:"client_id"`
	Secret         string `ini:"secret"`
	ApiUrl         string `ini:"api_url"`
	DefaultAccount string `ini:"default_account"`
	AuthUrl        string `ini:"auth_url"`
	Username       string `ini:"username"`
}

type Config struct {
	Conn          brightbox.Connection
	App           *kingpin.Application
	Clients       map[string]ConfigClient
	DefaultClient string
	Client        *ConfigClient
}

type TokenCacheSource struct {
	ts oauth2.TokenSource
}

func (s TokenCacheSource) Token() (*oauth2.Token, error) {
	t, e := s.ts.Token()
	return t, e
}

func (c *Config) Configure() error {
	c.Conn.ClientID = c.Client.ClientID
	c.Conn.ApiUrl = c.Client.ApiUrl
	c.Conn.ClientSecret = c.Client.Secret
	return nil
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	cfg.Clients = make(map[string]ConfigClient)
	err := cfg.readConfig()
	if err != nil {
		return cfg, err
	}
	return cfg, err
}

func (c *Config) readConfig() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	cfg, err := ini.Load(path.Join(u.HomeDir, ".brightbox/config"))
	if err != nil {
		return err
	}
	//cfg.BlockMode = false
	core := cfg.Section("core")
	c.DefaultClient = core.Key("default_client").String()
	for _, sec := range cfg.Sections() {
		if sec.Name() != "DEFAULT" && sec.Name() != "core" {
			cs := new(ConfigClient)
			cs.ClientName = sec.Name()
			if cs.ClientName == "" {
				continue
			}
			err = sec.MapTo(cs)
			if err != nil {
				continue
			}
			c.Clients[cs.ClientName] = *cs
		}
	}
	err = c.setClient(c.DefaultClient)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) setClient(clientName string) error {
	if clientName == "" {
		return nil
	}
	client, e := c.Clients[clientName]
	if e {
		c.Client = &client
		//c.Conn.CachedToken = &oauth2.Token{AccessToken: "1540e27f393a77b704c789eb40e7940f76bda1c9"}
		return nil
	} else {
		return fmt.Errorf("client '%s' not found in config.", clientName)
	}
}

func NewConfigAndConfigure(clientName string) (*Config, error) {
	cfg, err := NewConfig()
	if err != nil {
		return cfg, err
	}
	err = cfg.setClient(clientName)
	if err != nil {
		return cfg, err
	}
	err = cfg.Configure()
	if err != nil {
		return cfg, err
	}
	err = cfg.Conn.Connect()
	return cfg, err
}

type ConfigCommand struct {
	All  bool
	Json bool
	Id   string
}

func (l *ConfigCommand) list(pc *kingpin.ParseContext) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "NAME", "CLIENTID", "SECRET", "API_URL", "AUTH_URL")
	var keys sort.StringSlice
	for key, _ := range cfg.Clients {
		keys = append(keys, key)
	}
	sort.Sort(keys)
	for _, key := range keys {
		c := cfg.Clients[key]
		if cfg.DefaultClient == key {
			key = "*" + key
		}
		listRec(w, key, c.ClientID, c.Secret,
			c.ApiUrl, c.AuthUrl)
	}
	return nil
}

func (l *ConfigCommand) show(pc *kingpin.ParseContext) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	c := cfg.Clients[l.Id]
	if err != nil {
		return err
	}
	drawShow(w, []interface{}{
		"name", c.ClientName,
		"default", cfg.DefaultClient == c.ClientName,
		"client_id", c.ClientID,
		"api_url", c.ApiUrl,
		"auth_url", c.AuthUrl,
		"username", c.Username,
		"secret", c.Secret,
		"default_account", c.DefaultAccount,
	})
	return nil

}

func ConfigureServersCommand(app *kingpin.Application) {
	c := &ConfigCommand{}
	configcmd := app.Command("config", "manage cli configuration")
	configcmd.Command("list", "list local client configurations").Action(c.list)
	show := configcmd.Command("show", "view details on a client config").Action(c.show)
	show.Arg("name", "name or id of client config").Required().StringVar(&c.Id)
}
