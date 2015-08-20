package cli

import (
	"../brightbox"
	"encoding/base64"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"os"
	"io/ioutil"
	"fmt"
)

type ServersCommand struct {
	App        *CliApp
	All        bool
	Id         string
	IdList     []string
	ImageId    string
	Name       string
	ServerType string
	Zone       string
	Groups     string
	UserData   string
	UserDataFile *os.File
	Base64     bool
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

func (l *ServersCommand) create(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	newServer := brightbox.CreateServerOptions{
		Image: l.ImageId,
		Name:  l.Name,
	}

	if l.Zone != "" {
		zoneId, err := l.App.Client.resolveZoneId(l.Zone)
		if err != nil {
			return err
		}
		newServer.Zone = zoneId
	}

	if l.ServerType != "" {
		typeId, err := l.App.Client.resolveServerTypeId(l.ServerType)
		if err != nil {
			return err
		}
		newServer.ServerType = typeId
	}

	if l.Groups != "" {
		groups := strings.Split(l.Groups, ",")
		if len(groups) > 1 || (len(groups) == 1 && groups[0] != "") {
			newServer.ServerGroups = groups
		}
	}

	var userData []byte
	if l.UserData != "" {
		userData = []byte(l.UserData)
	}
	if l.UserDataFile != nil {
		defer l.UserDataFile.Close()
		fi, err := l.UserDataFile.Stat()
		if err != nil {
			return err
		}
		if fi.Size() > 2<<13 {
			return fmt.Errorf("User data file cannot exceed 16k")
		}
		userData, err = ioutil.ReadAll(l.UserDataFile)
		if err != nil {
			return err
		}
	}

	if len(userData) > 0 && l.Base64 {
		newServer.UserData = base64.StdEncoding.EncodeToString(userData)
	} else if len(userData) > 0 {
		newServer.UserData = string(userData)
	}
	if len(newServer.UserData) > 2<<13 {
		return fmt.Errorf("User data cannot exceed 16k")
	}

	server, err := l.App.Client.CreateServer(&newServer)
	if err != nil {
		return err
	}
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	s := server
	listRec(
		w, s.Id, s.Status, s.ServerType.Handle,
		s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
		s.Image.Id, collectById(s.CloudIPs), s.Name)
	return nil

}

func (l *ServersCommand) destroy(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	for _, id := range l.IdList {
		fmt.Printf("Destroying server %s\n", id)
		err := l.App.Client.DestroyServer(id)
		if err != nil {
			l.App.Errorf("%s: %s", err.Error(), id)
		}
	}
	return nil
}

func ConfigureServersCommand(app *CliApp) {
	cmd := ServersCommand{App: app}
	servers := app.Command("servers", "Manage cloud servers")
	servers.Command("list", "List cloud servers").Action(cmd.list)
	show := servers.Command("show", "View details on a cloud server").Action(cmd.show)
	show.Arg("identifier", "Identifier of server to show").Required().StringVar(&cmd.Id)
	create := servers.Command("create", "Create a new cloud server").Action(cmd.create)
	create.Arg("image identifier", "Identifier of image with which to create the server").Required().StringVar(&cmd.ImageId)
	create.Flag("name", "Name to give the new server").Short('n').StringVar(&cmd.Name)
	create.Flag("type", "Server type for the new server").Short('t').StringVar(&cmd.ServerType)
	create.Flag("zone", "Availability zone in which to place the new server").Short('z').StringVar(&cmd.Zone)
	create.Flag("groups", "The server groups to which the new server should belong. Comma separate multiple groups.").Short('g').StringVar(&cmd.Groups)
	create.Flag("user-data", "Specify the user data as a string").PlaceHolder("USERDATA").StringVar(&cmd.UserData)
	create.Flag("user-data-file", "Specify the user data from local file").PlaceHolder("FILENAME").OpenFileVar(&cmd.UserDataFile, 0, 0)
	create.Flag("base64", "Base64 encode the user data (default: true)").Default("true").BoolVar(&cmd.Base64)
	destroy := servers.Command("destroy", "Destroy a cloud server").Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of server to destroy").Required().StringsVar(&cmd.IdList)
}
