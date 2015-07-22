package brightbox

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type ApiError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

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
	TokenUrl     string
	TokenConfig  *oauth2.Config
	Token        string
	Scopes       []string
	Client       *http.Client
	ctx          context.Context
}

type Server struct {
	Resource
	Name              string
	Status            string
	Locked            bool
	Hostname          string
	Fqdn              string
	CreatedAt         *time.Time `json:"created_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	ServerType        ServerType `json:"server_type"`
	CompatabilityMode bool       `json:"compatability_mode"`
	Zone              Zone
	Image             Image
	CloudIPs          []CloudIP `json:"cloud_ips"`
	Interfaces        []ServerInterface
}

type ServerInterface struct {
	Resource
	MacAddress  string `json:"mac_address"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
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
	if c.ApiUrl == "" {
		c.ApiUrl = "https://api.gb1.brightbox.com"
	}
	if c.TokenUrl == "" {
		c.TokenUrl = c.ApiUrl + "/token"
	}
	if c.ClientConfig == nil {
		c.ClientConfig = &clientcredentials.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			TokenURL:     c.TokenUrl,
			Scopes:       c.Scopes,
		}
	}
	if c.TokenConfig == nil {
		c.TokenConfig = &oauth2.Config{}
	}
	c.ctx = oauth2.NoContext
}

func (c *Connection) Connect() error {
	var client *http.Client
	c.setDefaults()
	if c.Token != "" {
		client = oauth2.NewClient(c.ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.Token}))
	} else if c.ClientConfig != nil {
		client = c.ClientConfig.Client(c.ctx)
	}
	if client == nil {
		return errors.New("Failed to create oauth2 client")
	}
	c.Client = client
	return nil
}

func DisplayIds(resources interface{}) string {
	val := reflect.ValueOf(resources)
	if val.Kind() == reflect.Slice {
		var ids = make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
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
	u, _ = u.Parse("/1.0" + path)
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *Connection) MakeApiRequest(method string, path string) (*[]byte, error) {
	var body []byte
	var apierror ApiError
	res, err := c.Client.Get(c.api_url(path))
	if err != nil {
		return nil, err
	}
	body, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 200 {
		return &body, err
	} else {
		json.Unmarshal(body, &apierror)
		return nil, errors.New(apierror.ErrorDescription)
	}
}

func (c *Connection) Servers() (*[]Server, *[]byte, error) {
	var servers []Server
	var body *[]byte
	body, err := c.MakeApiRequest("get", "/servers")
	if err != nil {
		return nil, body, err
	}
	err = json.Unmarshal(*body, &servers)
	if err != nil {
		return &servers, body, err
	}
	return &servers, body, err
}

func (c *Connection) Server(identifier string) (*Server, *[]byte, error) {
	var server Server
	var body *[]byte
	body, err := c.MakeApiRequest("get", "/servers/"+identifier)
	if err != nil {
		return nil, body, err
	}
	err = json.Unmarshal(*body, &server)
	if err != nil {
		return &server, body, err
	}
	return &server, body, err
}
