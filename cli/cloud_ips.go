package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type CloudIPsCommand struct {
	*CliApp
	All          bool
	Id           string
	IdList       []string
	ImageId      string
	Name         string
	CloudIPType  string
	Zone         string
	Groups       string
	UserData     string
	UserDataFile *os.File
	Base64       bool
}

func cloudIPDestinationId(cip *brightbox.CloudIP) string {
	if cip.Server != nil {
		return cip.Server.Id
	}
	return ""
}

func (l *CloudIPsCommand) list(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	CloudIPs, err := l.Client.CloudIPs()
	if err != nil {
		return err
	}
	listRec(w, "ID", "STATUS", "PUBLIC_IP", "DESTINATION", "REVERSEDNS", "PTS", "NAME")
	for _, s := range *CloudIPs {
		listRec(
			w, s.Id, s.Status, s.PublicIP,
			cloudIPDestinationId(&s), s.ReverseDns,
			len(s.PortTranslators), s.Name)
	}
	return nil
}

func (l *CloudIPsCommand) show(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	s, err := l.Client.CloudIP(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
	}

	drawShow(w, []interface{}{
		"id", s.Id,
		"name", s.Name,
		"status", s.Status,
		"public_ip", s.PublicIP,
		"fqdn", nil,
		"reverse_dns", s.ReverseDns,
		"destination", cloudIPDestinationId(s),
		"port_translators", s.PortTranslators,
	})
	return nil

}

func (l *CloudIPsCommand) destroy(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	for _, id := range l.IdList {
		fmt.Printf("Destroying Cloud IP %s\n", id)
		err := l.Client.DestroyCloudIP(id)
		if err != nil {
			l.Errorf("%s: %s", err.Error(), id)
		}
	}
	return nil
}

func ConfigureCloudIPsCommand(app *CliApp) {
	cmd := CloudIPsCommand{CliApp: app}
	cloudips := app.Command("cloudips", "Manage Cloud IPs")
	cloudips.Command("list", "List Cloud IPs").Action(cmd.list)
	show := cloudips.Command("show", "View details on a Cloud IP").Action(cmd.show)
	show.Arg("identifier", "Identifier of Cloud IP to show").Required().StringVar(&cmd.Id)
	destroy := cloudips.Command("destroy", "Destroy a Cloud IP").Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of Cloud IP to destroy").Required().StringsVar(&cmd.IdList)
}
