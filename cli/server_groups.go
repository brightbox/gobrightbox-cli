package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServerGroupsCommand struct {
	App         *CliApp
	Id          string
	IdList      []string
	Name        *string
	Description *string
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
	listRec(w, "ID", "SERVER_COUNT", "FWPOLICY", "NAME")
	for _, s := range *groups {
		listRec(
			w, s.Id, len(s.Servers), s.FirewallPolicy.Id, s.Name)
	}
	return nil
}

func (l *ServerGroupsCommand) show(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	g, err := l.App.Client.ServerGroup(l.Id)
	if err != nil {
		l.App.Fatalf(err.Error())
	}

	drawShow(w, []interface{}{
		"id", g.Id,
		"name", g.Name,
		"default", g.Default,
		"servers", collectById(g.Servers),
		"firewall_policy", g.FirewallPolicy.Id,
		"description", g.Description,
	})
	return nil

}

func (l *ServerGroupsCommand) create(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	newGroup := brightbox.ServerGroupOptions{
		Name:        l.Name,
		Description: l.Description,
	}

	group, err := l.App.Client.CreateServerGroup(&newGroup)
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "SERVER_COUNT", "FWPOLICY", "NAME")
	listRec(
		w, group.Id, len(group.Servers), group.FirewallPolicy.Id, group.Name)
	return nil
}

func (l *ServerGroupsCommand) update(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	updateGroup := brightbox.ServerGroupOptions{
		Identifier:  l.Id,
		Name:        l.Name,
		Description: l.Description,
	}
	group, err := l.App.Client.UpdateServerGroup(&updateGroup)
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "SERVER_COUNT", "FWPOLICY", "NAME")
	listRec(
		w, group.Id, len(group.Servers), group.FirewallPolicy.Id, group.Name)
	return nil
}

func (l *ServerGroupsCommand) destroy(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Destroying server group %s\n", id)
		err := l.App.Client.DestroyServerGroup(id)
		if err != nil {
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func ConfigureServerGroupsCommand(app *CliApp) {
	cmd := ServerGroupsCommand{App: app}
	groups := app.Command("groups", "manage server groups")
	groups.Command("list", "list server groups").
		Default().Action(cmd.list)

	show := groups.Command("show", "View details of a server group").
		Action(cmd.show)
	show.Arg("identifier", "id of server group to show").
		StringVar(&cmd.Id)

	create := groups.Command("create", "Create a new server group").
		Action(cmd.create)
	create.Flag("name", "Name to give the new server group").
		Short('n').SetValue(&pStringValue{&cmd.Name})
	create.Flag("description", "Description to give the new server group").
		Short('d').SetValue(&pStringValue{&cmd.Description})

	update := groups.Command("update", "Update a new server group").
		Action(cmd.update)
	update.Arg("identifier", "id of server group to update").
		Required().StringVar(&cmd.Id)
	update.Flag("name", "Set a new name for the server group").
		Short('n').SetValue(&pStringValue{&cmd.Name})
	update.Flag("description", "Set a new description for the server group").
		Short('d').SetValue(&pStringValue{&cmd.Description})

	destroy := groups.Command("destroy", "Destroy a server group").
		Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of server groupto destroy").
		Required().StringsVar(&cmd.IdList)

}
