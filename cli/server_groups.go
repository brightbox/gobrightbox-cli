package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServerGroupsCommand struct {
	App *CliApp
	Id  string
}

func (l *ServerGroupsCommand) list(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	groups, err := l.App.Client.ServerGroups()
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "SERVER_COUNT", "NAME")
	for _, s := range *groups {
		listRec(
			w, s.Id, len(s.Servers), s.Name)
	}
	return nil
}

func ConfigureServerGroupsCommand(app *CliApp) {
	cmd := ServerGroupsCommand{App: app}
	groups := app.Command("groups", "manage server groups")
	groups.Command("list", "list server groups").Action(cmd.list)
}
