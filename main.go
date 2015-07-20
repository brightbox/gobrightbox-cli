package main

import (
	"./brightbox"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"text/tabwriter"
)

func tabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}

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

func configureServersCommand(app *kingpin.Application) {
	c := &ServersCommand{}
	ls := app.Command("servers", "manage servers").Action(c.run)
	ls.Flag("all", "List all servers.").Short('a').BoolVar(&c.All)
}

func main() {
	app := kingpin.New("brightbox", "Bleh")
	configureServersCommand(app)
	//kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.1").Author("John Leach")
	kingpin.MustParse(app.Parse(os.Args[1:]))
	//kingpin.Parse()
	//fmt.Printf("%v, %s\n", *verbose, *name)
}
