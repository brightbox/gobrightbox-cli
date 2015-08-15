package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type CliApp struct {
	*kingpin.Application
	ClientName string
	AccountId string
}

func New() *CliApp {
	a := new(CliApp)
	a.Application = kingpin.New("brightbox", "Bleh")
	a.Flag("client", "client to authenticate with.").OverrideDefaultFromEnvar("CLIENT").StringVar(&a.ClientName)
	a.Flag("account", "id of account to limit queries to").OverrideDefaultFromEnvar("ACCOUNT").StringVar(&a.AccountId)

	ConfigureServersCommand(a)
	ConfigureConfigCommand(a)
	ConfigureAccountsCommand(a)
	return a
}
