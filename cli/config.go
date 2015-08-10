package cli

import (
	"../brightbox"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"sort"
	"net/url"
)

// Represents a Client section from the config
// Can also be used as a TokenSource for oauth2 transport
type Client struct {
	ClientName     string
	ClientID       string `ini:"client_id"`
	Secret         string `ini:"secret"`
	ApiUrl         string `ini:"api_url"`
	DefaultAccount string `ini:"default_account"`
	AuthUrl        string `ini:"auth_url"`
	Username       string `ini:"username"`
	tokenCache     *TokenCacher
	client         *brightbox.Client
}

func (c *Client) findAuthUrl() (string) {
	var err error
	var u *url.URL
	if c.AuthUrl != "" {
		u, err = url.Parse(c.AuthUrl)
	}
	if u == nil || err != nil {
		u, err = url.Parse(c.ApiUrl)
	}
	if u == nil || err != nil {
		return ""
	}
	rel, _ := url.Parse("/token")
	u = u.ResolveReference(rel)
	if u == nil || err != nil {
		return ""
	}
	return u.String()
}

func (c *Client) Token() (*oauth2.Token, error) {
	if c.tokenCache == nil {
		c.tokenCache = &TokenCacher{Key: c.ClientName}
	}
	token := c.tokenCache.Read()
	if token != nil {
		return token, nil
	}
	oc := clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.Secret,
		TokenURL:     c.findAuthUrl(),
		Scopes:       []string{},
	}
	token, err := oc.Token(oauth2.NoContext)
	if err != nil {
		return nil, err
	}
	c.tokenCache.Write(token)
	return c.tokenCache.Read(), err
}

type Config struct {
	App           *kingpin.Application
	Clients       map[string]Client
	DefaultClient string
	Client        *Client
}

type TokenCacher struct {
	Key   string
	token *oauth2.Token
}

func (tc *TokenCacher) Read() *oauth2.Token {
	if tc.token != nil && tc.token.Valid() == true {
		return tc.token
	}
	filename := tc.jsonFilename()
	if filename == nil {
		return nil
	}
	token_json, err := ioutil.ReadFile(*filename)
	if err != nil {
		return nil
	}
	var token oauth2.Token
	err = json.Unmarshal(token_json, &token)
	if err != nil {
		return nil
	}
	if token.Valid() == false {
		tc.Clear()
		return nil
	}
	tc.token = &token
	return tc.token
}

func (tc *TokenCacher) jsonFilename() *string {
	dir := configDirectory()
	if dir == nil {
		return nil
	}
	filename := path.Join(*dir, tc.Key+".oauth_token.json")
	return &filename
}

func (tc *TokenCacher) Write(token *oauth2.Token) {
	if token == nil {
		return
	}
	// FIXME: make sure token differs from one we already have
	tc.token = token
	filename := tc.jsonFilename()
	if filename == nil {
		return
	}
	j, err := json.Marshal(token)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(*filename, j, 0600)
}

func (tc *TokenCacher) Clear() {
	tc.token = nil
	filename := tc.jsonFilename()
	if filename == nil {
		return
	}
	os.Remove(*filename)
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	cfg.Clients = make(map[string]Client)
	err := cfg.readConfig()
	if err != nil {
		return cfg, err
	}
	return cfg, err
}

func configDirectory() *string {
	u, err := user.Current()
	if err != nil {
		return nil
	}
	dir := path.Join(u.HomeDir, ".brightbox")
	return &dir
}

func (c *Config) readConfig() error {
	config_dir := configDirectory()
	if config_dir == nil {
		return nil
	}
	config_filename := path.Join(*config_dir, "config")
	cfg, err := ini.Load(config_filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	//cfg.BlockMode = false
	core := cfg.Section("core")
	c.DefaultClient = core.Key("default_client").String()
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
	tc := oauth2.NewClient(oauth2.NoContext, cfg.Client)
	apiUrl, err := url.Parse(cfg.Client.ApiUrl)
	cfg.Client.client = brightbox.NewClient(*apiUrl, tc)
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
