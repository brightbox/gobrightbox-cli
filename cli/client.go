package cli

import (
	"github.com/brightbox/gobrightbox"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

// Represents a Client section from the config
// Can also be used as a TokenSource for oauth2 transport
type Client struct {
	*brightbox.Client `json:"-"`
	ClientName     string `json:"-"`
	ClientID       string `json:"client_id,omitempty"`
	Secret         string `json:"secret,omitempty"`
	ApiUrl         string `json:"api_url,omitempty"`
	DefaultAccount string `json:"default_account,omitempty"`
	AuthUrl        string `json:"auth_url,omitempty"`
	Username       string `json:"username,omitempty"`
	tokenCache     *TokenCacher
	tokenSource    oauth2.TokenSource
}

func (c *Client) TokenCache() *TokenCacher {
	if c.tokenCache == nil {
		c.tokenCache = &TokenCacher{Key: c.ClientName}
	}
	return c.tokenCache
}

func (c *Client) Setup(accountId string) error {
	tc := oauth2.NewClient(oauth2.NoContext, c.TokenSource())
	if accountId == "" {
		accountId = c.DefaultAccount
	}
	client, err := brightbox.NewClient(c.ApiUrl, accountId, tc)
	if err != nil {
		return err
	}
	c.Client = client
	return nil
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
