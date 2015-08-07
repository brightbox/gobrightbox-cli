package brightbox

import (
	"errors"
	"golang.org/x/oauth2"
	"net/http"
)

type TokenCacher interface {
	Read() *oauth2.Token
	Write(token *oauth2.Token)
	Clear()
}

type AuthOptions struct {
	ApiUrl       string
	UserName     string
	UserSecret   string
	AccountId    string
	ClientID     string
	ClientSecret string
	TokenCache   TokenCacher
}

type Transport struct {
	BrightboxAuth *AuthOptions
	O2Transport   *oauth2.Transport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.O2Transport.RoundTrip(req)
	if res != nil {
		//fmt.Print(res)
		if res.StatusCode == 401 {
			if t.BrightboxAuth.TokenCache != nil {
				t.BrightboxAuth.TokenCache.Clear()
			}
		}
	}
	return res, err
}

func (t *Transport) CancelRequest(req *http.Request) {
	t.O2Transport.CancelRequest(req)
}

func (a *AuthOptions) Token() (*oauth2.Token, error) {
	var token *oauth2.Token
	if a.TokenCache != nil {
		token = a.TokenCache.Read()
		a.TokenCache.Write(token)
	}
	if token == nil {
		return &oauth2.Token{}, errors.New("couldn't obtain an auth token")
	}
	return token, nil
}

func (a *AuthOptions) NewClient() (*http.Client, error) {

	client := &http.Client{
		Transport: &Transport{
			BrightboxAuth: a,
			O2Transport: &oauth2.Transport{
				Source: a,
			},
		},
	}

	return client, nil
}
