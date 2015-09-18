package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type AccountsCommand struct {
	App *CliApp
	Id  string
}

func (l *AccountsCommand) list(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	accounts, err := l.App.Client.Accounts()
	if err != nil {
		return err
	}
	listRec(w, "ID", "RAM_USED", "ROLE", "NAME")
	for _, a := range *accounts {
		listRec(
			w, a.Id, a.RamUsed, "",
			a.Name)
	}
	return nil
}

func ConfigureAccountsCommand(app *CliApp) {
	cmd := AccountsCommand{App: app}
	accounts := app.Command("accounts", "manage accounts")
	accounts.Command("list", "list accounts").Action(cmd.list)
}
