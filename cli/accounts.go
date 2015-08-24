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

func (l *AccountsCommand) show(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	s, err := l.App.Client.Server(l.Id)
	if err != nil {
		l.App.Fatalf(err.Error())
	}

	drawShow(w, []interface{}{
		"id", s.Id,
		"status", s.Status,
		"locked", s.Locked,
		"name", s.Name,
		"created_at", s.CreatedAt,
		"deleted_at", s.DeletedAt,
		"zone", s.Zone.Handle,
		"type", s.ServerType.Id,
		"type_name", s.ServerType.Name,
		"type_handle", s.ServerType.Handle,
		"ram", s.ServerType.Ram,
		"cores", s.ServerType.Cores,
		"disk", s.ServerType.DiskSize,
		"compatability_mode", s.CompatabilityMode,
		"image", s.Image.Id,
		"image_name", s.Image.Name,
		"arch", s.Image.Arch,
		"private_ips", collectByField(s.Interfaces, "IPv4Address"),
		"cloud_ips", collectByField(s.CloudIPs, "PublicIP"),
		"ipv6_ips", collectByField(s.Interfaces, "IPv6Address"),
		"cloud_ip_ids", collectByField(s.CloudIPs, "Id"),
		"hostname", s.Hostname,
		"fqdn", s.Fqdn,
		"public_hostname", "public." + s.Fqdn,
		"ipv6_hostname", "ipv6." + s.Fqdn,
		"snapshots", collectById(s.Snapshots),
		"server_groups", collectById(s.ServerGroups),
	})
	return nil

}

func ConfigureAccountsCommand(app *CliApp) {
	cmd := AccountsCommand{App: app}
	accounts := app.Command("accounts", "manage accounts")
	accounts.Command("list", "list accounts").Action(cmd.list)
	show := accounts.Command("show", "view details of an account").Action(cmd.show)
	show.Arg("identifier", "identifier of server to show").Required().StringVar(&cmd.Id)
}
