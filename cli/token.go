package cli

import (
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

type TokenCommand struct {
	*CliApp
	Id     string
	Force  bool
	Format string
}

func (l *TokenCommand) create(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()

	if l.Force {
		l.Client.TokenCache().Clear()
	}
	token, err := l.Client.TokenSource().Token()
	if token == nil {
		l.Fatalf("No cached OAuth token found for %s", l.ClientName)
	}
	switch l.Format {
	case "json":
		err := json.NewEncoder(w).Encode(token)
		if err != nil {
			return err
		}
	case "text":
		drawShow(w, []interface{}{
			"access_token", token.AccessToken,
			"token_type", token.TokenType,
			"expiry", token.Expiry,
		})
	case "curl":
		fmt.Fprintf(w, "curl -H 'Authorization: Bearer %s' %s\n", token.AccessToken, l.Client.ApiUrl)
	}

	return nil
}

func (l *TokenCommand) clear(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	l.Client.TokenCache().Clear()
	return nil
}

func ConfigureTokenCommand(app *CliApp) {
	cmd := TokenCommand{CliApp: app}
	token := app.Command("token", "manage oauth tokens")
	create := token.Command("create", "return a valid token for the client, create one if necessary").Action(cmd.create)
	create.Flag("clear", "clear the local cache first and create a new token").BoolVar(&cmd.Force)
	create.Flag("format", "the output format: text, json or curl").Default("text").EnumVar(&cmd.Format, "text", "json", "curl")
	token.Command("clear", "clear the local token cache for this client").Action(cmd.clear)
}
