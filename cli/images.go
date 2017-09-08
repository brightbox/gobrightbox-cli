package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"sort"
	"strings"
)

var (
	defaultImageListFields = []string{"id", "owner", "type", "created", "status", "arch", "name"}
	defaultImageShowFields = []string{"id", "type", "owner", "created_at", "status", "locked", "arch",
		"name", "description", "username", "virtual_size", "disk_size", "public", "compatibility_mode",
		"official", "ancestor_id", "licence_name"}
)

type imagesCommand struct {
	*CLIApp
	Id      string
	IdList  []string
	ShowAll bool
	Fields  string
}

func imageFields(i *brightbox.Image) map[string]string {
	owner := i.Owner
	itype := i.SourceType
	if i.Official {
		owner = "brightbox"
		itype = "official"
	}
	name := i.Name
	if i.CompatibilityMode {
		name += " (compat)"
	}
	name += " (" + i.Arch + ")"
	var status string
	if i.Status != "available" {
		status = i.Status
	} else if i.Public {
		status = "public"
	} else {
		status = "private"
	}

	return map[string]string{
		"id":                 i.Id,
		"type":               itype,
		"owner":              owner,
		"created":            i.CreatedAt.Format("2006-01-02"),
		"created_at":         formatTime(&i.CreatedAt),
		"status":             status,
		"locked":             formatBool(i.Locked),
		"arch":               i.Arch,
		"name":               i.Name,
		"description":        i.Description,
		"username":           i.Username,
		"virtual_size":       formatInt(i.VirtualSize),
		"disk_size":          formatInt(i.DiskSize),
		"public":             formatBool(i.Public),
		"compatibility_mode": formatBool(i.CompatibilityMode),
		"official":           formatBool(i.Official),
		"ancestor_id":        i.AncestorId,
		"licence_name":       i.LicenceName,
	}
}

type imagesForDisplay []brightbox.Image

func (il imagesForDisplay) Len() int      { return len(il) }
func (il imagesForDisplay) Swap(i, j int) { il[i], il[j] = il[j], il[i] }
func (il imagesForDisplay) Less(i, j int) bool {
	a := il.sortKeys(i)
	b := il.sortKeys(j)
	for e := 0; e < len(a); e++ {
		switch a[e].(type) {
		case string:
			if a[e].(string) != b[e].(string) {
				return a[e].(string) < b[e].(string)
			}
		case int64:
			if a[e].(int64) != b[e].(int64) {
				return a[e].(int64) < b[e].(int64)
			}
		case bool:
			if a[e].(bool) != b[e].(bool) {
				return a[e].(bool) == false && b[e].(bool) == true
			}
		}
	}
	return false
}

func (il imagesForDisplay) sortKeys(i int) []interface{} {
	im := il[i]
	return []interface{}{
		!im.Official,
		im.SourceType == "snapshot",
		im.Name, im.Arch,
		im.Status == "deprecated",
		!im.Public,
		-im.CreatedAt.Unix(),
	}
}

func (l *imagesCommand) list(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}

	// Get the account id in parallel to everything else, we don't need it until
	// later anyway.
	achan := make(chan string)
	go func() {
		achan <- l.accountId()
	}()

	out := new(RowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))

	images, err := l.Client.Images()
	if err != nil {
		return err
	}
	sortedImages := make(imagesForDisplay, 0, len(images))
	for _, i := range images {
		sortedImages = append(sortedImages, i)
	}
	sort.Sort(sortedImages)

	accountId := <-achan
	out.SendHeader()
	for _, i := range sortedImages {
		if l.ShowAll == false {
			if !i.Official && i.Owner != accountId {
				continue
			}
		}
		if err = out.Write(imageFields(&i)); err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func (l *imagesCommand) show(pc *kingpin.ParseContext) error {

	err := l.Configure()
	if err != nil {
		return err
	}
	out := new(ShowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))

	i, err := l.Client.Image(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
	}
	if err = out.Write(imageFields(i)); err != nil {
		return err
	}
	out.Flush()
	return nil

}

func (l *imagesCommand) destroy(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	returnError := false
	for _, id := range l.IdList {
		fmt.Printf("Destroying image %s\n", id)
		err := l.Client.DestroyImage(id)
		if err != nil {
			l.Errorf("%s: %s", err.Error(), id)
			returnError = true
		}
	}
	if returnError {
		return errGeneric
	}
	return nil
}

func configureImagesCommand(app *CLIApp) {
	cmd := imagesCommand{CLIApp: app}
	images := app.Command("images", "Manage server images")
	list := images.Command("list", "List server images").Default().Action(cmd.list)
	list.Flag("fields", "Which fields to display").
		Default(strings.Join(defaultImageListFields, ",")).
		StringVar(&cmd.Fields)
	list.Flag("show-all", "Show all public images from all accounts").Default("false").BoolVar(&cmd.ShowAll)
	show := images.Command("show", "View details of a server image").Action(cmd.show)
	show.Flag("fields", "Which fields to display").
		Default(strings.Join(defaultImageShowFields,",")).
		StringVar(&cmd.Fields)
	show.Arg("identifier", "Identifier of image to show").Required().StringVar(&cmd.Id)
	destroy := images.Command("destroy", "Destroy a server image").Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of image to destroy").Required().StringsVar(&cmd.IdList)

}
