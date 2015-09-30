package cli

import (
	"encoding/json"
	"fmt"
	"github.com/casimir/xdg-go"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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

func (c *Config) Clients() (clients []Client, err error) {
	filenames, err := filepath.Glob(xdgapp.ConfigPath("*.client.json"))
	if err != nil {
		return
	}
	for _, f := range filenames {
		f = path.Base(strings.TrimSuffix(f, ".client.json"))
		client, err := c.Client(f)

		if err == nil {
			c.clients[f] = *client
			clients = append(clients, *client)
		}
	}
	return clients, nil
}

func (c *Config) SaveClientConfig(client *Client) error {
	if client == nil {
		panic("Can't save client config for nil client")
	}
	j, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return err
	}
	filename := xdgapp.ConfigPath(client.ClientName + ".client.json")
	err = ioutil.WriteFile(filename, j, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Client(cname string) (*Client, error) {
	client, exists := c.clients[cname]
	if exists == true {
		return &client, nil
	}
	filename := xdgapp.ConfigPath(cname + ".client.json")
	jd, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("client '%s' not found in config.", cname)
	}
	if err != nil {
		return nil, err
	}
	client = Client{}
	err = json.Unmarshal(jd, &client)
	if err != nil {
		return nil, err
	}
	client.ClientName = cname
	c.clients[cname] = client
	return &client, nil
}

func (c *Config) CurrentClient() *Client {
	return c.currentClient
}

func (c *Config) DefaultClient() *Client {
	client, err := c.Client(c.defaultClientName)
	if err != nil && client == nil {
		clients, err := c.Clients()
		if err != nil {
			return nil
		}
		if len(clients) > 0 {
			client = &clients[0]
		}
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
	err = c.ReadGlobal()
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) ReadGlobal() error {
	filename := xdgapp.ConfigPath("cli.json")
	jd, err := ioutil.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		return nil
	}

	globalConf := make(map[string]string)
	err = json.Unmarshal(jd, &globalConf)
	if err != nil {
		return err
	}
	c.defaultClientName = globalConf["default_client"]
	return nil
}

func (c *Config) WriteGlobal() error {
	filename := xdgapp.ConfigPath("cli.json")

	globalConf := make(map[string]string)
	globalConf["default_client"] = c.defaultClientName
	jd, err := json.MarshalIndent(globalConf, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, jd, 0600)
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
	clients, err := cfg.Clients()
	if err != nil {
		return err
	}
	dc := cfg.DefaultClient()
	for _, c := range clients {
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
