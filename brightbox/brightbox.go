package brightbox

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Client struct {
	Auth       *AuthOptions
	HttpClient *http.Client
}

func (c *Client) New(a *AuthOptions) error {
	c.Auth = a
	hc, err := a.NewClient()
	if err != nil {
		return err
	}
	c.HttpClient = hc
	return nil
}

type ApiError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Resource struct {
	Id string
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

func (c *Client) api_url(path string) string {
	u, _ := url.Parse(c.Auth.ApiUrl)
	v := u.Query()
	if c.Auth.AccountId != "" {
		v.Set("account_id", c.Auth.AccountId)
	}
	u, _ = u.Parse("/1.0" + path)
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *Client) MakeApiRequest(method string, path string) (*[]byte, error) {
	var body []byte
	var apierror ApiError
	res, err := c.HttpClient.Get(c.api_url(path))
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

func (c *Client) Servers() (*[]Server, *[]byte, error) {
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

func (c *Client) Server(identifier string) (*Server, *[]byte, error) {
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
