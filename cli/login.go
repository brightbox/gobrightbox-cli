package cli

import (
	"fmt"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

type LoginCommand struct {
	App   *CliApp
	Email string
}

func (l *LoginCommand) login(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	if l.App.ClientName == "" {
		err = l.App.Config.SetClient(l.Email)
		if err != nil {
			return err
		}
	}
	w := tabWriterRight()
	defer w.Flush()

	oc := l.App.Client.oauthConfig()
	var token *oauth2.Token
	switch oc := oc.(type) {
	case oauth2.Config:
		var password string
		fmt.Printf("Password for %s: ", l.App.Client.Username)
		fmt.Scanln(&password)
		token, err = oc.PasswordCredentialsToken(oauth2.NoContext, l.App.Client.Username, password)
		if err != nil {
			l.App.Fatalf("%s", err)
		}
	default:
		l.App.Fatalf("Client config %s isn't for password authentication", l.App.Client.ClientName)
	}
	l.App.Client.TokenCache().Write(token)
	return nil
}

func ConfigureLoginCommand(app *CliApp) {
	cmd := LoginCommand{App: app}
	login := app.Command("login", "Authenticate with user credentials").Action(cmd.login)
	login.Arg("email address", "Your user's email address").
		Required().StringVar(&cmd.Email)
}
