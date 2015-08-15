package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServersCommand struct {
	App        *CliApp
	All        bool
	Id         string
}

func (l *ServersCommand) list(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	servers, err := l.App.Client.Servers()
	if err != nil {
		return err
	}
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	for _, s := range *servers {
		listRec(
			w, s.Id, s.Status, s.ServerType.Handle,
			s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
			s.Image.Id, collectById(s.CloudIPs), s.Name)
	}
	return nil
}

func (l *ServersCommand) show(pc *kingpin.ParseContext) error {
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

func ConfigureServersCommand(app *CliApp) {
	cmd := ServersCommand{App: app}
	servers := app.Command("servers", "manage cloud servers")
	servers.Command("list", "list cloud servers").Action(cmd.list)
	show := servers.Command("show", "view details on cloud servers").Action(cmd.show)
	show.Arg("identifier", "identifier of server to show").Required().StringVar(&cmd.Id)
}
