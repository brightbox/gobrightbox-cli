package cli

import (
	"fmt"
	"github.com/casimir/xdg-go"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
	"os"
)

var (
	xdgapp = xdg.App{Name: "brightbox"}
)

type Config struct {
	App               *kingpin.Application
	defaultClientName string
	currentClient     *Client
	clients           map[string]Client
}

func NewConfig() (*Config, error) {
	c := new(Config)
	err := c.Setup()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) SaveClientConfig(client *Client) error {
	if client == nil {
		panic("Can't save client config for nil client")
	}

	filename := xdgapp.ConfigPath("config")
	cfg, err := ini.Load(filename)
	if os.IsNotExist(err) {
		cfg = ini.Empty()
	} else if err != nil {
		return err
	}

	section := cfg.Section(client.ClientName)
	section.Key("client_id").SetValue(client.ClientID)
	section.Key("secret").SetValue(client.Secret)
	section.Key("api_url").SetValue(client.ApiUrl)
	section.Key("auth_url").SetValue(client.AuthUrl)
	section.Key("default_account").SetValue(client.DefaultAccount)
	section.Key("username").SetValue(client.Username)
	err = cfg.SaveTo(filename)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Client(cname string) (*Client, error) {
	client, exists := c.clients[cname]
	if exists == false {
		return nil, fmt.Errorf("client '%s' not found in config.", cname)
	}
	return &client, nil
}

func (c *Config) CurrentClient() *Client {
	return c.currentClient
}

func (c *Config) DefaultClient() *Client {
	client, err := c.Client(c.defaultClientName)
	if err != nil && client == nil {
		return nil
	}
	return client
}

func (c *Config) Setup() error {
	err := os.MkdirAll(xdgapp.ConfigPath(""), 0750)
	if err != nil {
		return err
	}
	err = os.MkdirAll(xdgapp.CachePath(""), 0750)
	if err != nil {
		return err
	}
	if c.clients == nil {
		c.clients = make(map[string]Client)
	}
	err = c.Read()
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Read() error {
	filename := xdgapp.ConfigPath("config")
	cfg, err := ini.Load(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	core := cfg.Section("core")
	c.defaultClientName = core.Key("default_client").String()
	for _, sec := range cfg.Sections() {
		if sec.Name() != "DEFAULT" && sec.Name() != "core" {
			cs := new(Client)
			cs.ClientName = sec.Name()
			if cs.ClientName == "" {
				continue
			}
			err = sec.MapTo(cs)
			if err != nil {
				continue
			}
			c.clients[cs.ClientName] = *cs
			if c.defaultClientName == "" {
				c.defaultClientName = cs.ClientName
			}

		}
	}
	return nil

}

func (c *Config) Write() error {
	filename := xdgapp.ConfigPath("config")
	cfg, err := ini.Load(filename)
	if os.IsNotExist(err) {
		cfg = ini.Empty()
	} else if err != nil {
		return err
	}

	section := cfg.Section("core")
	key := section.Key("default_client")
	key.SetValue(c.defaultClientName)
	err = cfg.SaveTo(filename)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetClient(clientName string) error {
	if clientName == "" {
		c.currentClient = c.DefaultClient()
		return nil
	}
	client, err := c.Client(clientName)
	if err == nil && client != nil {
		c.currentClient = client
		return nil
	} else {
		return fmt.Errorf("client '%s' not found in config.", clientName)
	}
}

type ConfigCommand struct {
	Id string
}

func (l *ConfigCommand) list(pc *kingpin.ParseContext) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "NAME", "CLIENTID", "SECRET", "API_URL", "AUTH_URL")
	dc := cfg.DefaultClient()
	for _, c := range cfg.clients {
		name := c.ClientName
		if dc != nil && dc.ClientName == name {
			name = "*" + name
		}
		listRec(w, name, c.ClientID, c.Secret,
			c.ApiUrl, c.findAuthUrl())
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
	c, err := cfg.Client(l.Id)
	if err != nil {
		return err
	}
	dc := cfg.DefaultClient()
	drawShow(w, []interface{}{
		"name", c.ClientName,
		"default", dc != nil && dc.ClientName == c.ClientName,
		"client_id", c.ClientID,
		"api_url", c.ApiUrl,
		"auth_url", c.AuthUrl,
		"username", c.Username,
		"secret", c.Secret,
		"default_account", c.DefaultAccount,
	})
	return nil

}

func ConfigureConfigCommand(app *CliApp) {
	c := &ConfigCommand{}
	configcmd := app.Command("config", "manage cli configuration")
	configcmd.Command("list", "list local client configurations").
		Default().Action(c.list)
	show := configcmd.Command("show", "view details on a client config").Action(c.show)
	show.Arg("name", "name or id of client config").Required().StringVar(&c.Id)
}
