package cli

import (
	"errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

var (
	genericError = errors.New("Errors were encountered")
)

type CliApp struct {
	*kingpin.Application
	ClientName string
	AccountId  string
	Config     *Config
	Client     *Client
}

func New() *CliApp {
	a := new(CliApp)
	a.Application = kingpin.New("brightbox", "Bleh")
	a.Flag("client", "client to authenticate with.").OverrideDefaultFromEnvar("CLIENT").StringVar(&a.ClientName)
	a.Flag("account", "id of account to limit queries to").OverrideDefaultFromEnvar("ACCOUNT").StringVar(&a.AccountId)

	ConfigureServersCommand(a)
	ConfigureConfigCommand(a)
	ConfigureAccountsCommand(a)
	ConfigureServerGroupsCommand(a)
	ConfigureTokenCommand(a)
	ConfigureImagesCommand(a)
	ConfigureCloudIPsCommand(a)
	ConfigureEventsCommand(a)
	ConfigureLoginCommand(a)
	return a
}

func (c *CliApp) Configure() error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	c.Config = cfg

	clientName := c.ClientName
	if clientName == "" {
		clientName = cfg.defaultClientName
	}
	if clientName == "" {
		return nil
	}
	err = cfg.SetClient(clientName)
	if err != nil {
		return err
	}
	err = cfg.CurrentClient().Setup(c.AccountId)
	if err != nil {
		return err
	}
	c.Client = cfg.CurrentClient()
	return nil
}

// Try to get an account id for the connection, either as specified in the
// config or by looking up the api client id
func (c *CliApp) accountId() string {
	if c.AccountId != "" {
		return c.AccountId
	}
	if strings.HasPrefix(c.Client.ClientID, "cli-") {
		apiClient, err := c.Client.ApiClient(c.Client.ClientID)
		if err == nil {
			return apiClient.Account.Id
		}
	}
	return ""
}
