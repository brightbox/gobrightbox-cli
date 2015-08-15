package brightbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL   *url.URL
	client    *http.Client
	UserAgent string
}

func NewClient(apiUrl url.URL, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{
		client:  httpClient,
		BaseURL: &apiUrl,
	}
	return c
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	return req, nil
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
	CompatabilityMode bool       `json:"compatibility_mode"`
	Zone              Zone
	Image             Image
	CloudIPs          []CloudIP `json:"cloud_ips"`
	Interfaces        []ServerInterface
	Snapshots         []Image
	ServerGroups      []ServerGroup `json:"server_groups"`
}

type ServerGroup struct {
	Resource
	Name        string
	CreatedAt   *time.Time `json:"created_at"`
	Description string
	Default     bool
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
	PublicIP   string `json:"public_ip"`
	ReverseDns string `json:"reverse_dns"`
	Name       string
}

func (c *Client) MakeApiRequest(method string, path string, reqbody interface{}, resbody interface{}) (*http.Response, error) {
	var body []byte
	req, err := c.NewRequest(method, path, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		if resbody != nil {
			err := json.NewDecoder(res.Body).Decode(resbody)
			if err != nil {
				return res, err
			}
		}
		return res, err
	} else {
		apierr := new(ApiError)
		json.Unmarshal(body, apierr)
		return res, fmt.Errorf("%s: %s %s", res.Status, res.Request.URL.String(),apierr.ErrorDescription)
	}
}

func (c *Client) Servers() (*[]Server, error) {
	servers := new([]Server)
	_, err := c.MakeApiRequest("GET", "/1.0/servers", nil, servers)
	if err != nil {
		return nil, err
	}
	return servers, err
}

func (c *Client) Server(identifier string) (*Server, error) {
	server := new(Server)
	_, err := c.MakeApiRequest("GET", "/1.0/servers/"+identifier, nil, server)
	if err != nil {
		return nil, err
	}
	return server, err
}
