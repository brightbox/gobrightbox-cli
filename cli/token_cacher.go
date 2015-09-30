package cli

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
)

type TokenCacher struct {
	Key   string
	token *oauth2.Token
}

func (tc *TokenCacher) Read() *oauth2.Token {
	if tc.token != nil {
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
	tc.token = &token
	return tc.token
}

func (tc *TokenCacher) jsonFilename() *string {
	filename := xdgapp.CachePath(tc.Key+".oauth_token.json")
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
