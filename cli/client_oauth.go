package cli

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

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
