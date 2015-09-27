package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

type ServerGroupsCommand struct {
	App         *CliApp
	Id          string
	Dst         string
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
		Id:          l.Id,
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

func (l *ServerGroupsCommand) add(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	fmt.Printf("Adding servers %s to server group %s\n", strings.Join(l.IdList, ", "), l.Id)
	_, err = l.App.Client.AddServersToServerGroup(l.Id, l.IdList)
	if err != nil {
		return err
	}
	return nil
}

func (l *ServerGroupsCommand) remove(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	fmt.Printf("Removing servers %s to server group %s\n", strings.Join(l.IdList, ", "), l.Id)
	_, err = l.App.Client.RemoveServersFromServerGroup(l.Id, l.IdList)
	if err != nil {
		return err
	}
	return nil
}

func (l *ServerGroupsCommand) move(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	fmt.Printf("Moving servers %s from server group %s to server group %s\n", strings.Join(l.IdList, ", "), l.Dst, l.Id)
	_, err = l.App.Client.MoveServersToServerGroup(l.Id, l.Dst, l.IdList)
	if err != nil {
		return err
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

	add := groups.Command("add_servers", "Add servers to a server group").
		Action(cmd.add)
	add.Arg("group_identifier", "Identifier of group to add the servers to").
		Required().StringVar(&cmd.Id)
	add.Arg("server_identifiers", "Identifiers of servers to add to the group").
		Required().StringsVar(&cmd.IdList)

	rem := groups.Command("remove_servers", "Remove servers from a server group").
		Action(cmd.remove)
	rem.Arg("group_identifier", "Identifier of group to remove the servers from").
		Required().StringVar(&cmd.Id)
	rem.Arg("server_identifiers", "Identifiers of servers to remove from the group").
		Required().StringsVar(&cmd.IdList)

	mv := groups.Command("move_servers", "Move servers between server groups").
		Action(cmd.move)
	mv.Arg("src_group_identifier", "Identifier of group to move the servers from").
		Required().StringVar(&cmd.Id)
	mv.Arg("dst_group_identifier", "Identifier of group to move the servers to").
		Required().StringVar(&cmd.Dst)
	mv.Arg("server_identifiers", "Identifiers of servers to move").
		Required().StringsVar(&cmd.IdList)

}
