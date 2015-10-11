package cli

import (
	"fmt"
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"sort"
)

type ImagesCommand struct {
	*CliApp
	Id      string
	IdList  []string
	ShowAll bool
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

func (l *ImagesCommand) list(pc *kingpin.ParseContext) error {
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

	w := tabWriter()
	defer w.Flush()

	images, err := l.Client.Images()
	if err != nil {
		return err
	}
	sortedImages := make(imagesForDisplay, 0, len(*images))
	for _, i := range *images {
		sortedImages = append(sortedImages, i)
	}
	sort.Sort(sortedImages)

	accountId := <-achan

	listRec(w, "ID", "OWNER", "TYPE", "CREATED", "STATUS", "SIZE", "NAME")
	for _, i := range sortedImages {
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
		if l.ShowAll == false {
			if !i.Official && i.Owner != accountId {
				continue
			}
		}
		listRec(
			w, i.Id, owner, itype,
			i.CreatedAt.Format("2006-01-02"),
			status, i.VirtualSize, name)
	}
	return nil
}

func (l *ImagesCommand) show(pc *kingpin.ParseContext) error {

	err := l.Configure()
	if err != nil {
		return err
	}
	w := tabWriterRight()
	defer w.Flush()
	s, err := l.Client.Image(l.Id)
	if err != nil {
		l.Fatalf(err.Error())
	}

	owner := s.Owner
	itype := s.SourceType
	if s.Official {
		owner = "brightbox"
		itype = "official"
	}
	drawShow(w, []interface{}{
		"id", s.Id,
		"type", itype,
		"owner", owner,
		"created_at", s.CreatedAt,
		"status", s.Status,
		"locked", s.Locked,
		"arch", s.Arch,
		"name", s.Name,
		"description", s.Description,
		"username", s.Username,
		"virtual_size", s.VirtualSize,
		"disk_size", s.DiskSize,
		"public", s.Public,
		"compatibility_mode", s.CompatibilityMode,
		"official", s.Official,
		"ancestor_id", s.AncestorId,
		"license_name", s.LicenseName,
	})
	return nil

}

func (l *ImagesCommand) destroy(pc *kingpin.ParseContext) error {
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
		return genericError
	}
	return nil
}

func ConfigureImagesCommand(app *CliApp) {
	cmd := ImagesCommand{CliApp: app}
	images := app.Command("images", "Manage server images")
	list := images.Command("list", "List server images").Default().Action(cmd.list)
	list.Flag("show-all", "Show all public images from all accounts").Default("false").BoolVar(&cmd.ShowAll)
	show := images.Command("show", "View details of a server image").Action(cmd.show)
	show.Arg("identifier", "Identifier of image to show").Required().StringVar(&cmd.Id)
	destroy := images.Command("destroy", "Destroy a server image").Action(cmd.destroy)
	destroy.Arg("identifier", "Identifier of image to destroy").Required().StringsVar(&cmd.IdList)

}
