package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

type ServersCommand struct {
	All        bool
	Id         string
	ClientName string
}

func (l *ServersCommand) list(pc *kingpin.ParseContext) error {
	cfg, err := NewConfigAndConfigure(l.ClientName)
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	servers, err := cfg.Client.client.Servers()
	if err != nil {
		return err
	}
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	for _, s := range *servers {
		listRec(
			w, s.Id, s.Status, s.ServerType.Handle,
			s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
			s.Image.Id, DisplayIds(s.CloudIPs), s.Name)
	}
	return nil
}

func (l *ServersCommand) show(pc *kingpin.ParseContext) error {
	cfg, err := NewConfigAndConfigure(l.ClientName)
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	s, err := cfg.Client.client.Server(l.Id)
	if err != nil {
		return err
	}

	private_ips := make([]string, len(s.Interfaces))
	ipv6_ips := make([]string, len(s.Interfaces))
	for i, iface := range s.Interfaces {
		private_ips[i] = iface.IPv4Address
		ipv6_ips[i] = iface.IPv6Address
	}
	cloud_ips := make([]string, len(s.CloudIPs))
	cloud_ip_ids := make([]string, len(s.CloudIPs))
	for i, cip := range s.CloudIPs {
		cloud_ips[i] = cip.PublicIP
		cloud_ip_ids[i] = cip.Id
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
		"private_ips", strings.Join(private_ips, ", "),
		"cloud_ips", strings.Join(cloud_ips, ", "),
		"ipv6_ips", strings.Join(ipv6_ips, ", "),
		"cloud_ip_ids", strings.Join(cloud_ip_ids, ", "),
		"hostname", s.Hostname,
		"fqdn", s.Fqdn,
		"public_hostname", "public." + s.Fqdn,
		"ipv6_hostname", "ipv6." + s.Fqdn,
		"snapshots", DisplayIds(s.Snapshots),
		"server_groups", DisplayIds(s.ServerGroups),
	})
	return nil

}

func ConfigureConfigCommand(app *kingpin.Application) {
	cmd := new(ServersCommand)
	servers := app.Command("servers", "manage cloud servers")
	servers.Command("list", "list cloud servers").Action(cmd.list)
	show := servers.Command("show", "view details on cloud servers").Action(cmd.show)
	show.Arg("identifier", "identifier of server to show").Required().StringVar(&cmd.Id)
	app.Flag("client", "client to authenticate with.").StringVar(&cmd.ClientName)
}
