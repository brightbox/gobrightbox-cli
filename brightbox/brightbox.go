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
	AccountId string
}

type ApiError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Resource struct {
	Id string
}

func NewClient(apiUrl url.URL, accountId *string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{
		client:  httpClient,
		BaseURL: &apiUrl,
	}
	if accountId != nil {
		c.AccountId = *accountId
	}
	return c
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	if c.AccountId != "" {
		q := u.Query()
		q.Set("account_id", c.AccountId)
		u.RawQuery = q.Encode()
	}

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
		return res, fmt.Errorf("%s: %s %s", res.Status, res.Request.URL.String(), apierr.ErrorDescription)
	}
}
