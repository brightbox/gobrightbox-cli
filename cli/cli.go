package cli

import (
	"errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

var (
	errGeneric = errors.New("Errors were encountered")
)

// CLIApp represents a cli application instance
type CLIApp struct {
	*kingpin.Application
	ClientName string
	AccountId  string
	Config     *config
	Client     *Client
}

// New initializes the brightbox cli application
func New() *CLIApp {
	a := new(CLIApp)
	a.Application = kingpin.New("brightbox", "Bleh")
	a.Flag("client", "client to authenticate with.").OverrideDefaultFromEnvar("CLIENT").StringVar(&a.ClientName)
	a.Flag("account", "id of account to limit queries to").OverrideDefaultFromEnvar("ACCOUNT").StringVar(&a.AccountId)

	configureServersCommand(a)
	configureConfigCommand(a)
	configureAccountsCommand(a)
	configureServerGroupsCommand(a)
	configureTokenCommand(a)
	configureImagesCommand(a)
	configureCloudIPsCommand(a)
	configureEventsCommand(a)
	configureLoginCommand(a)
	return a
}

func (c *CLIApp) Configure() error {
	cfg, err := newConfig()
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
	err = cfg.setClient(clientName)
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
func (c *CLIApp) accountId() string {
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
