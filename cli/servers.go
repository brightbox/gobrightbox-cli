package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"../brightbox"
)

// Context for "ls" command
type ServersCommand struct {
	All bool
}

func (l *ServersCommand) run(c *kingpin.ParseContext) error {
	conn := &brightbox.Connection{
		Token:     os.Getenv("BRIGHTBOX_TOKEN"),
		AccountId: os.Getenv("BRIGHTBOX_ACCOUNT"),
		ApiUrl:    os.Getenv("BRIGHTBOX_API_URL"),
	}
	conn.Connect()

	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	for _, s := range conn.Servers() {
		listRec(
			w, s.Id, s.Status, s.ServerType.Handle,
			s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
			s.Image.Id, brightbox.DisplayIds(s.CloudIPs), s.Name)
	}
	return nil
}

func ConfigureServersCommand(app *kingpin.Application) {
	c := &ServersCommand{}
	ls := app.Command("servers", "manage servers").Action(c.run)
	ls.Flag("all", "List all servers.").Short('a').BoolVar(&c.All)
}
