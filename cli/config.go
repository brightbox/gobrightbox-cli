package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
	"net/url"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"
)

// Represents a Client section from the config
// Can also be used as a TokenSource for oauth2 transport
type Client struct {
	*brightbox.Client
	ClientName     string
	ClientID       string `ini:"client_id"`
	Secret         string `ini:"secret"`
	ApiUrl         string `ini:"api_url"`
	DefaultAccount string `ini:"default_account"`
	AuthUrl        string `ini:"auth_url"`
	Username       string `ini:"username"`
	tokenCache     *TokenCacher
	tokenSource    oauth2.TokenSource
}

func (c *Client) TokenCache() *TokenCacher {
	if c.tokenCache == nil {
		c.tokenCache = &TokenCacher{Key: c.ClientName}
	}
	return c.tokenCache
}

func (c *Client) findAuthUrl() string {
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

func (c *Client) findRegionDomain() string {
	u, err := url.Parse(c.ApiUrl)
	if err != nil {
		return ""
	}
	if strings.HasPrefix(u.Host, "api.") {
		return strings.TrimPrefix(u.Host, "api.")
	}
	return ""
}

// Token returns the cached OAuth token from disk if it's still valid, or
// retrieves a new one from the token source.
func (c *Client) Token() (*oauth2.Token, error) {
	token := c.TokenCache().Read()
	if token != nil && token.Valid() == true {
		return token, nil
	} else {
		if c.tokenSource == nil {
			panic(fmt.Sprintf("No tokenSource set up yet for %s", c.ClientName))
		}
		token, err := c.tokenSource.Token()
		if err != nil {
			return nil, err
		}

		c.TokenCache().Write(token)
		return token, nil
	}
}

// Return an appropriate oauth2 config for this Client
func (c *Client) oauthConfig() interface{} {
	if c.Username != "" {
		return oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.Secret,
			Endpoint: oauth2.Endpoint{
				TokenURL: c.findAuthUrl(),
			},
		}
	} else {
		return clientcredentials.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.Secret,
			TokenURL:     c.findAuthUrl(),
			Scopes:       []string{},
		}
	}
}

// Setup the OAuth token source, which can then issue (and cache) OAuth
// tokens. API clients can always be used to get new tokens. Password auth
// credentials need a valid refresh token, or they error out (and need a login
// to get a new refresh token).
func (c *Client) TokenSource() oauth2.TokenSource {
	if c.tokenSource != nil {
		return c
	}
	oc := c.oauthConfig()
	switch oc := oc.(type) {
	case oauth2.Config:
		c.tokenSource = oc.TokenSource(oauth2.NoContext, c.TokenCache().Read())
	case clientcredentials.Config:
		c.tokenSource = oc.TokenSource(oauth2.NoContext)
	}
	return c
}

type Config struct {
	App           *kingpin.Application
	Clients       map[string]Client
	DefaultClient string
	Client        *Client
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
	err = c.SetClient(c.DefaultClient)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetClient(clientName string) error {
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

func NewConfigAndConfigure(clientName string, accountId string) (*Config, error) {
	cfg, err := NewConfig()
	if err != nil {
		return cfg, err
	}
	err = cfg.SetClient(clientName)
	if err != nil {
		return cfg, err
	}
	tc := oauth2.NewClient(oauth2.NoContext, cfg.Client.TokenSource())
	if accountId == "" {
		accountId = cfg.Client.DefaultAccount
	}
	cfg.Client.Client, err = brightbox.NewClient(cfg.Client.ApiUrl, accountId, tc)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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

func ConfigureConfigCommand(app *CliApp) {
	c := &ConfigCommand{}
	configcmd := app.Command("config", "manage cli configuration")
	configcmd.Command("list", "list local client configurations").Action(c.list)
	show := configcmd.Command("show", "view details on a client config").Action(c.show)
	show.Arg("name", "name or id of client config").Required().StringVar(&c.Id)
}
