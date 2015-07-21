package cli

import (
	"../brightbox"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Context for "ls" command
type ServersCommand struct {
	All  bool
	Json bool
}

func (l *ServersCommand) list(pc *kingpin.ParseContext) error {
	c, err := NewConfigAndConnect()
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
	c, err := NewConfig()
	if err != nil {
		return err
	}
	err = c.Conn.Connect()
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	s, body, err := c.Conn.Server("srv-aop8h")
	if err != nil {
		return err
	}
	if l.Json {
		fmt.Fprint(w, PrettyPrintJson(*body))
	} else {
		listRec(w, "Identifier", s.Id)
		listRec(w, "Status", s.Status)
	}
	return nil

}

func ConfigureServersCommand(app *kingpin.Application) {
	c := &ServersCommand{}
	list := app.Command("servers", "manage cloud servers").Action(c.list)
	show := app.Command("show", "view details on cloud servers").Action(c.show)
	list.Flag("all", "List all servers.").Short('a').BoolVar(&c.All)
	list.Flag("json", "Output raw json from server.").BoolVar(&c.Json)
	show.Flag("json", "Output raw json from server.").BoolVar(&c.Json)
}
