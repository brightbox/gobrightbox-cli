package brightbox

import (
	"encoding/json"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Resource struct {
	Id string
}

type Connection struct {
	ApiUrl       string
	UserName     string
	UserSecret   string
	AccountId    string
	UserAgent    string
	ClientID     string
	ClientSecret string
	ClientConfig *clientcredentials.Config
	TokenConfig  *oauth2.Config
	Token        string
	Client       *http.Client
	ctx          context.Context
}

type Server struct {
	Resource
	Name       string
	Status     string
	Locked     bool
	Hostname   string
	Fqdn       string
	CreatedAt  time.Time  `json:"created_at"`
	ServerType ServerType `json:"server_type"`
	Zone       Zone
	Image      Image
	CloudIPs   []CloudIP `json:"cloud_ips"`
}

type ServerType struct {
	Resource
	Name     string
	Status   string
	Handle   string
	Cores    int
	Ram      int
	DiskSize int `json:"disk_size"`
}

type Zone struct {
	Resource
	Handle string
}

type Image struct {
	Resource
	Name        string
	Username    string
	Status      string
	Locked      bool
	Description string
	Source      string
	Arch        string
	CreatedAt   time.Time `json:"created_at"`
	Official    bool
	Public      bool
	Owner       string
}

type CloudIP struct {
	Resource
	Status     string
	PublicIP   string
	ReverseDns string
	Name       string
}

func (c *Connection) setDefaults() {
	if c.ClientConfig == nil {
		c.ClientConfig = &clientcredentials.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			TokenURL:     "https://api.gb1s.brightbox.com/token",
			Scopes:       []string{""},
		}
	}
	if c.TokenConfig == nil {
		c.TokenConfig = &oauth2.Config{}
	}
	c.ctx = oauth2.NoContext
}

func (c *Connection) Connect() {
	c.setDefaults()
	//client := c.ClientConfig.Client(c.ctx)
	client := oauth2.NewClient(c.ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token, TokenType: "OAuth"}))
	if client != nil {
		c.Client = client
	}
}

func DisplayIds(resources interface{}) string {
	val := reflect.ValueOf(resources)
	if val.Kind() == reflect.Slice {
		var ids = make([]string,val.Len())
		for i:=0;i<val.Len(); i++ {
			rval := val.Index(i)
			if rval.Kind() == reflect.Struct && rval.FieldByName("Id").IsValid() {
				ids[i] = rval.FieldByName("Id").String()
			}
		}
		return strings.Join(ids, ",")
	}
	return ""
}

func (c *Connection) api_url(path string) string {
	u, _ := url.Parse(c.ApiUrl)
	v := u.Query()
	if c.AccountId != "" {
		v.Set("account_id", c.AccountId)
	}
	u, _ = u.Parse("/1.0/" + path)
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *Connection) Servers() []Server {
	res, err := c.Client.Get(c.api_url("/servers"))
	if err != nil {
		log.Fatal(err)
	}
	var servers []Server
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &servers)
	return servers
	//log.Print(servers)
}
