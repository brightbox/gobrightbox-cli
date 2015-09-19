package cli

import (
	"encoding/base64"
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

type ServersCommand struct {
	App               *CliApp
	All               bool
	Id                string
	IdList            []string
	ImageId           string
	Name              *string
	ServerType        string
	Zone              string
	Groups            *string
	UserData          *string
	UserDataFile      *os.File
	Base64            bool
	CompatibilityMode *bool
}

func (l *ServersCommand) list(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}

	w := tabWriter()
	defer w.Flush()
	var groupFilter []string
	if l.Groups != nil {
		groupFilter = strings.Split(*l.Groups, ",")
	}
	servers, err := l.App.Client.Servers()
	if err != nil {
		return err
	}
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	for _, s := range *servers {
		if len(groupFilter) > 0 {
			matches := 0
			for _, gf := range groupFilter {
				for _, g := range s.ServerGroups {
					if g.Id == gf {
						matches += 1
					}
				}
			}
			if matches == 0 {
				continue
			}
		}
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
		"created_at", formatTime(s.CreatedAt),
		"deleted_at", formatTime(s.DeletedAt),
		"zone", s.Zone.Handle,
		"type", s.ServerType.Id,
		"type_name", s.ServerType.Name,
		"type_handle", s.ServerType.Handle,
		"ram", s.ServerType.Ram,
		"cores", s.ServerType.Cores,
		"disk", s.ServerType.DiskSize,
		"compatibility_mode", s.CompatibilityMode,
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
	newServer := brightbox.ServerOptions{
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

	if l.Groups != nil {
		groups := strings.Split(*l.Groups, ",")
		if len(groups) > 1 || (len(groups) == 1 && groups[0] != "") {
			newServer.ServerGroups = &groups
		}
	}

	var userData []byte
	if l.UserData != nil {
		userData = []byte(*l.UserData)
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
		bs := base64.StdEncoding.EncodeToString(userData)
		newServer.UserData = &bs
	} else if len(userData) > 0 {
		s := string(userData)
		newServer.UserData = &s
	}
	if newServer.UserData != nil && len(*newServer.UserData) > 2<<13 {
		return fmt.Errorf("User data cannot exceed 16k")
	}

	server, err := l.App.Client.CreateServer(&newServer)
	if err != nil {
		return err
	}
	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	s := server
	listRec(
		w, s.Id, s.Status, s.ServerType.Handle,
		s.Zone.Handle, s.CreatedAt.Format("2006-01-02"),
		s.Image.Id, collectById(s.CloudIPs), s.Name)
	return nil

}

func (l *ServersCommand) update(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	updateServer := brightbox.ServerOptions{Identifier: l.Id}

	updateServer.Name = l.Name

	if l.Groups != nil {
		groups := strings.Split(*l.Groups, ",")
		if len(groups) > 1 || (len(groups) == 1 && groups[0] != "") {
			updateServer.ServerGroups = &groups
		}
	}

	updateServer.CompatibilityMode = l.CompatibilityMode

	var userData []byte
	if l.UserData != nil {
		userData = []byte(*l.UserData)
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
		bs := base64.StdEncoding.EncodeToString(userData)
		updateServer.UserData = &bs
	} else if len(userData) > 0 {
		s := string(userData)
		updateServer.UserData = &s
	}
	if updateServer.UserData != nil && len(*updateServer.UserData) > 2<<13 {
		return fmt.Errorf("User data cannot exceed 16k")
	}

	w := tabWriter()
	defer w.Flush()
	listRec(w, "ID", "STATUS", "TYPE", "ZONE", "CREATED", "IMAGE", "CLOUDIPS", "NAME")
	returnError := false
	for _, id := range l.IdList {

		fmt.Printf("Updating server %s\n", id)
		updateServer.Identifier = id
		server, err := l.App.Client.UpdateServer(&updateServer)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
			continue
		}
		listRec(
			w, server.Id, server.Status, server.ServerType.Handle,
			server.Zone.Handle, server.CreatedAt.Format("2006-01-02"),
			server.Image.Id, collectById(server.CloudIPs), server.Name)
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) destroy(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Destroying server %s\n", id)
		err := l.App.Client.DestroyServer(id)
		if err != nil {
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) stop(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Stopping server %s\n", id)
		err := l.App.Client.StopServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) start(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Starting server %s\n", id)
		err := l.App.Client.StartServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) reboot(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Rebooting server %s\n", id)
		err := l.App.Client.RebootServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) reset(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Resetting server %s\n", id)
		err := l.App.Client.ResetServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) shutdown(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Shutting down server %s\n", id)
		err := l.App.Client.ShutdownServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) lock(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Locking server %s\n", id)
		err := l.App.Client.LockServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) unlock(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Unlocking server %s\n", id)
		err := l.App.Client.UnlockServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) snapshot(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Snapshotting server %s\n", id)
		img, err := l.App.Client.SnapshotServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
			continue
		}
		fmt.Printf("Snapsnot image %s started from server %s\n", img.Id, id)
	}
	if returnError {
		return genericError
	}
	return nil
}

func (l *ServersCommand) activateConsole(pc *kingpin.ParseContext) error {
	err := l.App.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Activating console for server %s\n", id)
		srv, err := l.App.Client.ActivateConsoleForServer(id)
		if err != nil {
			if len(l.IdList) == 1 {
				return err
			}
			l.App.Errorf("%s: %s", err.Error(), id)
			returnError = true
			continue
		}
		fmt.Printf("Console activated for server %s: %s\n", id, srv.FullConsoleUrl())
	}
	if returnError {
		return genericError
	}
	return nil
}

func ConfigureServersCommand(app *CliApp) {
	cmd := ServersCommand{App: app}
	servers := app.Command("servers", "Manage cloud servers")
	list := servers.Command("list", "List cloud servers").
		Default().Action(cmd.list)
	list.Flag("groups", "List only servers belonging to these groups").
		Short('g').SetValue(&pStringValue{&cmd.Groups})

	show := servers.Command("show", "View details on a cloud server").
		Action(cmd.show)
	show.Arg("identifier", "Identifier of server to show").
		Required().StringVar(&cmd.Id)

	create := servers.Command("create", "Create a new cloud server").
		Action(cmd.create)
	create.Arg("image identifier", "Identifier of image with which to create the server").
		Required().StringVar(&cmd.ImageId)
	create.Flag("name", "Name to give the new server").
		Short('n').SetValue(&pStringValue{&cmd.Name})
	create.Flag("type", "Server type for the new server").
		Short('t').StringVar(&cmd.ServerType)
	create.Flag("zone", "Availability zone in which to place the new server").
		Short('z').StringVar(&cmd.Zone)
	create.Flag("groups", "The server groups to which the new server should belong. Comma separate multiple groups.").
		Short('g').SetValue(&pStringValue{&cmd.Groups})
	create.Flag("user-data", "Specify the user data as a string").
		PlaceHolder("USERDATA").SetValue(&pStringValue{&cmd.UserData})
	create.Flag("user-data-file", "Specify the user data from local file").
		PlaceHolder("FILENAME").OpenFileVar(&cmd.UserDataFile, 0, 0)
	create.Flag("base64", "Base64 encode the user data (default: true)").
		Default("true").BoolVar(&cmd.Base64)

	update := servers.Command("update", "Update a cloud server").
		Action(cmd.update)
	update.Arg("identifier", "Identifier of servers to update").
		Required().StringsVar(&cmd.IdList)
	update.Flag("name", "Set a new name for the server").
		Short('n').SetValue(&pStringValue{&cmd.Name})
	update.Flag("groups", "Specify the groups to which the server should belong. Comma separate multiple groups.").
		Short('g').SetValue(&pStringValue{&cmd.Groups})
	update.Flag("user-data", "Specify the user data as a string").
		PlaceHolder("USERDATA").SetValue(&pStringValue{&cmd.UserData})
	update.Flag("user-data-file", "Specify the user data from local file").
		PlaceHolder("FILENAME").OpenFileVar(&cmd.UserDataFile, 0, 0)
	update.Flag("base64", "Base64 encode the user data (default: true)").
		Default("true").BoolVar(&cmd.Base64)
	update.Flag("compatibility-mode", "Enable/disable compatibility mode for the server").
		SetValue(&pBoolValue{&cmd.CompatibilityMode})

	destroy := servers.Command("destroy", "Destroy a cloud server").
		Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of server to destroy").
		Required().StringsVar(&cmd.IdList)

	stop := servers.Command("stop", "Stop a cloud server").
		Action(cmd.stop)
	stop.Arg("identifier", "Identifier of servers to stop").
		Required().StringsVar(&cmd.IdList)

	start := servers.Command("start", "Start a cloud server").
		Action(cmd.start)
	start.Arg("identifier", "Identifier of servers to start").
		Required().StringsVar(&cmd.IdList)

	reboot := servers.Command("reboot", "Reboot a cloud server").
		Action(cmd.reboot)
	reboot.Arg("identifier", "Identifier of servers to reboot").
		Required().StringsVar(&cmd.IdList)

	reset := servers.Command("reset", "Reset a cloud server").
		Action(cmd.reset)
	reset.Arg("identifier", "Identifier of servers to reset").
		Required().StringsVar(&cmd.IdList)

	shutdown := servers.Command("shutdown", "Shutdown a cloud server").
		Action(cmd.shutdown)
	shutdown.Arg("identifier", "Identifier of servers to shut down").
		Required().StringsVar(&cmd.IdList)

	lock := servers.Command("lock", "Lock a cloud server").
		Action(cmd.lock)
	lock.Arg("identifier", "Identifier of servers to lock").
		Required().StringsVar(&cmd.IdList)

	unlock := servers.Command("unlock", "Unlock a cloud server").
		Action(cmd.unlock)
	unlock.Arg("identifier", "Identifier of servers to unlock").
		Required().StringsVar(&cmd.IdList)

	snap := servers.Command("snapshot", "Snapshot a cloud server").
		Action(cmd.snapshot)
	snap.Arg("identifier", "Identifier of servers to snapshot").
		Required().StringsVar(&cmd.IdList)

	console := servers.Command("activate_console", "Activate the graphical console for a cloud server").
		Action(cmd.activateConsole)
	console.Arg("identifier", "Identifier of servers to snapshot").
		Required().StringsVar(&cmd.IdList)

}
