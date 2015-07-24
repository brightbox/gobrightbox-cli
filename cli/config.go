package cli

import (
	"../brightbox"
	"errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
	"os/user"
	"path"
	"sort"
)

type ConfigFileSection struct {
	SectionName    string
	ClientID       string `ini:"client_id"`
	Secret         string `ini:"secret"`
	ApiUrl         string `ini:"api_url"`
	DefaultAccount string `ini:"default_account"`
	AuthUrl        string `ini:"auth_url"`
	Username       string `ini:"username"`
	DefaultClient  string `ini:"default_client"`
}

type Config struct {
	Conn          *brightbox.Connection
	App           *kingpin.Application
	Clients       map[string]ConfigFileSection
	DefaultClient string
	Client        *ConfigFileSection
}

func (c *Config) Configure() error {
	/*c.Conn = &brightbox.Connection{
			Token:        os.Getenv("BRIGHTBOX_TOKEN"),
			AccountId:    os.Getenv("BRIGHTBOX_ACCOUNT"),
			ClientID:     os.Getenv("BRIGHTBOX_CLIENT_ID"),
			ClientSecret: os.Getenv("BRIGHTBOX_CLIENT_SECRET"),
			ApiUrl:       os.Getenv("BRIGHTBOX_API_URL"),
		}*/
	conn := brightbox.Connection{}
	conn.ClientID = c.Client.ClientID
	conn.ApiUrl = c.Client.ApiUrl
	conn.ClientSecret = c.Client.Secret
	c.Conn = &conn
	return nil
}

func NewConfig(clientName string) (*Config, error) {
	var config Config
	err := config.readConfig()
	if err != nil {
		return &config, err
	}
	err = config.setClient(clientName)
	if err != nil {
		return &config, err
	}

	err = config.Configure()

	return &config, err
}

func (c *Config) readConfig() error {
	if c.Clients == nil {
		c.Clients = make(map[string]ConfigFileSection)
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	cfg, err := ini.Load(path.Join(u.HomeDir, ".brightbox/config"))
	if err != nil {
		return err
	}
	cfg.BlockMode = false
	for _, sec := range cfg.Sections() {
		cs := new(ConfigFileSection)
		cs.SectionName = sec.Name()
		if cs.SectionName == "" {
			continue
		}
		err = sec.MapTo(cs)
		if err != nil {
			continue
		}
		if cs.SectionName == "DEFAULT" || cs.SectionName == "core" {
			c.DefaultClient = cs.DefaultClient
		} else {
			c.Clients[cs.SectionName] = *cs
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
		return nil
	} else {
		return errors.New("client name not found in config.")
	}
}

func NewConfigAndConnect(clientName string) (*Config, error) {
	config, err := NewConfig(clientName)
	if err != nil {
		return config, err
	}
	err = config.Conn.Connect()
	return config, err
}

type ConfigCommand struct {
	All  bool
	Json bool
	Id   string
}

func (l *ConfigCommand) list(pc *kingpin.ParseContext) error {
	cfg, err := NewConfig("")
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
	cfg, err := NewConfig("")
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
		"name", c.SectionName,
		"default", c.DefaultClient == c.SectionName,
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
