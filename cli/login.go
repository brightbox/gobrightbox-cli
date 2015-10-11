package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"github.com/howeyc/gopass"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

type LoginCommand struct {
	*CliApp
	Email          string
	ApiUrl         string
	AuthUrl        string
	ClientId       string
	Secret         string
	DefaultAccount string
}

func (l *LoginCommand) login(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}

	clientName := l.Email
	email := l.Email
	if strings.Contains(email, "/") {
		toks := strings.SplitN(email, "/", 2)
		email = toks[0]
	}
	var client *Client
	if l.ClientName == "" {
		client, err = l.Config.Client(clientName)
		if err != nil {
			client = new(Client)
		}
	} else {
		client, err = l.Config.Client(clientName)
		if err != nil {
			l.Fatalf("Couldn't find client config %s: %s", clientName, err)
		}
	}
	client.ClientName = clientName
	if l.ClientId != "" {
		client.ClientID = l.ClientId
	}
	if l.Secret != "" {
		client.Secret = l.Secret
	}
	if l.ApiUrl != "" {
		client.ApiUrl = l.ApiUrl
	}
	if l.AuthUrl != "" {
		client.AuthUrl = l.AuthUrl
	}
	if client.AuthUrl == "" && l.ApiUrl != "" {
		client.ApiUrl = l.ApiUrl
	}
	if l.Email != "" {
		client.Username = email
	}
	if l.DefaultAccount != "" {
		client.DefaultAccount = l.DefaultAccount
	}

	oc := client.oauthConfig()
	var token *oauth2.Token
	switch oc := oc.(type) {
	case oauth2.Config:
		fmt.Printf("Password for %s: ", client.Username)
		password := gopass.GetPasswd()
		if string(password) == "" {
			l.Fatalf("Password not provided.")
		}
		token, err = oc.PasswordCredentialsToken(oauth2.NoContext, client.Username, string(password))
		if err != nil {
			l.Fatalf("%s", err)
		}
	default:
		l.Fatalf("Client config %s isn't for password authentication", client.ClientName)
	}
	client.TokenCache().Write(token)

	// Choose a default account
	if client.DefaultAccount == "" {
		client.Setup("")
		accounts, err := client.Accounts()
		if err != nil {
			l.Errorf("Couldn't choose a default account: %s", err)
		}
		if len(*accounts) == 0 {
			l.Errorf("No accounts available to choose a default account")
		}
		var da brightbox.Account
		for _, a := range *accounts {
			if a.Status == "active" {
				if da.Id == "" {
					da = a
					continue
				}
				if a.RamUsed > da.RamUsed {
					da = a
				}
			}
		}
		if da.Id != "" {
			fmt.Printf("Selected account \"%s\" (%s) as default account\n", da.Name, da.Id)
			client.DefaultAccount = da.Id
		}
	}
	err = l.Config.SaveClientConfig(client)
	if err != nil {
		l.Fatalf("Couldn't save client config %s: %s", client.ClientName, err)
	}
	if l.Config.DefaultClient() == nil {
		l.Config.defaultClientName = client.ClientName
		l.Config.Write()
	}
	return nil
}

func ConfigureLoginCommand(app *CliApp) {
	cmd := LoginCommand{CliApp: app}
	login := app.Command("login", "Authenticate with user credentials").Action(cmd.login)
	login.Arg("email address", "Your user's email address").
		Required().StringVar(&cmd.Email)
	login.Flag("api-url", "url of Brightbox API").
		Default("https://api.gb1.brightbox.com").StringVar(&cmd.ApiUrl)
	login.Flag("auth-url", "url of Brightbox API authentication endpoint. Defaults to same as api-url.").
		StringVar(&cmd.AuthUrl)
	login.Flag("client-id", "OAuth client identifier to use").
		Default("app-12345").StringVar(&cmd.ClientId)
	login.Flag("secret", "OAuth client secret to use").
		Default("mocbuipbiaa6k6c").StringVar(&cmd.Secret)
	login.Flag("default-account", "Id of account to use by default with this client").
		StringVar(&cmd.DefaultAccount)

}