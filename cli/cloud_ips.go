package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
)

var (
	DefaultCloudIPListFields = []string{"id", "status", "public_ip", "pts", "reverse_dns", "name"}
	DefaultCloudIPShowFields = []string{"id", "name", "status", "public_ip", "fqdn", "reverse_dns", "destination", "port_translators"}
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
	Fields       string
}

func cloudIPDestinationId(cip *brightbox.CloudIP) string {
	if cip.Server != nil {
		return cip.Server.Id
	}
	return ""
}

func CloudIPFields(cip *brightbox.CloudIP) map[string]string {
	return map[string]string{
		"id":               cip.Id,
		"status":           cip.Status,
		"public_ip":        cip.PublicIP,
		"pts":              formatInt(len(cip.PortTranslators)),
		"reverse_dns":      cip.ReverseDns,
		"name":             cip.Name,
		"destination":      cloudIPDestinationId(cip),
		"fqdn":             "", // FIXME
		"port_translators": "", //FIXME
	}
}

func (l *CloudIPsCommand) list(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}

	out := new(RowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))
	out.SendHeader()

	cips, err := l.Client.CloudIPs()
	if err != nil {
		return err
	}
	for _, cip := range *cips {
		if err = out.Write(CloudIPFields(&cip)); err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func (l *CloudIPsCommand) show(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	out := new(ShowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))

	cip, err := l.Client.CloudIP(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
	}
	if err = out.Write(CloudIPFields(cip)); err != nil {
		return err
	}
	out.Flush()
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
	list := cloudips.Command("list", "List Cloud IPs").Action(cmd.list).Default()
	list.Flag("fields", "Which fields to display").
		Default(strings.Join(DefaultCloudIPListFields, ",")).
		StringVar(&cmd.Fields)

	show := cloudips.Command("show", "View details on a Cloud IP").Action(cmd.show)
	show.Flag("fields", "Which fields to display").
		Default(strings.Join(DefaultCloudIPShowFields, ",")).
		StringVar(&cmd.Fields)

	show.Arg("identifier", "Identifier of Cloud IP to show").Required().StringVar(&cmd.Id)
	destroy := cloudips.Command("destroy", "Destroy a Cloud IP").Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of Cloud IP to destroy").Required().StringsVar(&cmd.IdList)
}
