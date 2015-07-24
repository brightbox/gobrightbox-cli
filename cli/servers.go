package cli

import (
	"../brightbox"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServersCommand struct {
	All  bool
	Json bool
	Id   string
	ClientName string
}

func (l *ServersCommand) list(pc *kingpin.ParseContext) error {
	c, err := NewConfigAndConnect(l.ClientName)
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	servers, body, err := c.Conn.Servers()
	if err != nil {
		return err
	}
	if l.Json {
		if len(*servers) > 0 {
			fmt.Fprint(w, PrettyPrintJson(*body))
		}
	} else {
		listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
		for _, s := range *servers {
			listRec(
				w, s.Id, s.Status, s.ServerType.Handle,
				s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
				s.Image.Id, brightbox.DisplayIds(s.CloudIPs), s.Name)
		}
	}
	return nil
}

func (l *ServersCommand) show(pc *kingpin.ParseContext) error {
	c, err := NewConfigAndConnect(l.ClientName)
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	s, body, err := c.Conn.Server(l.Id)
	if err != nil {
		return err
	}
	if l.Json {
		fmt.Fprint(w, PrettyPrintJson(*body))
	} else {
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
			"private_ips", s.Interfaces[0].IPv4Address,
		})
	}
	return nil

}

func ConfigureConfigCommand(app *kingpin.Application) {
	c := &ServersCommand{}
	servers := app.Command("servers", "manage cloud servers")
	servers.Command("list", "list cloud servers").Action(c.list)
	show := servers.Command("show", "view details on cloud servers").Action(c.show)
	show.Arg("identifier", "identifier of server to show").Required().StringVar(&c.Id)
	app.Flag("json", "Output raw json from server.").BoolVar(&c.Json)
	app.Flag("client", "client to authenticate with.").StringVar(&c.ClientName)
}
