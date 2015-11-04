package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
	"time"
)

var (
	DefaultCloudIPListFields = []string{"id", "status", "public_ip", "pts", "destination", "name"}
	DefaultCloudIPShowFields = []string{"id", "name", "status", "public_ip", "fqdn", "reverse_dns", "destination", "port_translators"}
)

type CloudIPsCommand struct {
	*CliApp
	All          bool
	Id           string
	DestId       string
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
	Unmap        bool
}

func cloudIPDestinationId(cip *brightbox.CloudIP) string {
	if cip.Server != nil {
		return cip.Server.Id
	}
	if cip.Interface != nil {
		return cip.Interface.Id
	}

	if cip.ServerGroup != nil {
		return cip.Server.Id
	}
	if cip.DatabaseServer != nil {
		return cip.DatabaseServer.Id
	}
	if cip.LoadBalancer != nil {
		return cip.LoadBalancer.Id
	}
	return ""
}

func formatPortTranslators(pts []brightbox.PortTranslator) string {
	fpts := make([]string, len(pts))
	for i, pt := range pts {
		fpts[i] = fmt.Sprintf("%d:%d:%s", pt.Incoming, pt.Outgoing, pt.Protocol)
	}
	return strings.Join(fpts, ",")
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
		"fqdn":             cip.Fqdn,
		"port_translators": formatPortTranslators(cip.PortTranslators),
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
	for _, cip := range cips {
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

func (l *CloudIPsCommand) mapcip(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	cip, err := l.Client.CloudIP(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
	}
	if cip.Status == "mapped" && l.Unmap {
		fmt.Printf("Unmapping Cloud IP %s from %s\n", cip.Id, cloudIPDestinationId(cip))
		err = l.Client.UnMapCloudIP(cip.Id)
		if err != nil {
			l.Fatalf(err.Error())
		}
		for i := 2; i <= 6; i++ {
			cip, err = l.Client.CloudIP(l.Id)
			if err != nil {
				l.Fatalf(err.Error())
			}
			if cip.Status != "unmapped" {
				fmt.Printf("Cloud IP %s not yet unmapped, waiting longer %d more seconds.\n", cip.Id, i)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			}
		}
		if cip.Status != "unmapped" {
			l.Fatalf("Cloud IP %s wouldn't unmap, giving up.\n", cip.Id)
		}
	}
	fmt.Printf("Mapping Cloud IP %s to destination %s\n", cip.Id, l.DestId)
	err = l.Client.MapCloudIPtoServer(cip.Id, l.DestId)
	if err != nil {
		l.Fatalf("%s: %s to %s", err.Error(), cip.Id, l.DestId)
	}
	return nil
}

func (l *CloudIPsCommand) unmapcip(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	fmt.Printf("Unmapping Cloud IP %s\n", l.Id)
	err = l.Client.UnMapCloudIP(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
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

	mapcip := cloudips.Command("map", "Map a Cloud IP to another resource").Action(cmd.mapcip)
	mapcip.Arg("cloud-ip", "Identifier of the Cloud IP").Required().StringVar(&cmd.Id)
	mapcip.Arg("destination", "Identifier of the resource to which to map the Cloud IP").Required().StringVar(&cmd.DestId)

	mapcip.Flag("unmap", "Unmap any mapped Cloud IPs before remapping them").
		Default("false").
		BoolVar(&cmd.Unmap)

	unmapcip := cloudips.Command("unmap", "Unmap a mapped Cloud IP").Action(cmd.unmapcip)
	unmapcip.Arg("cloud-ip", "Identifier of the Cloud IP").Required().StringVar(&cmd.Id)

}
