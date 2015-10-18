package cli

import (
	"github.com/brightbox/gobrightbox"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

var (
	DefaultAccountListFields = []string{"id", "status", "role", "ram_used", "lb_used", "dbs_ram_used", "name"}
	DefaultAccountShowFields = []string{"id", "name", "status", "role", "cloud_ips_limit",
		"cloud_ips_used", "ram_limit", "ram_used", "lb_limit", "lb_used",
		"dbs_ram_limit", "dbs_ram_used", "library_ftp_host", "library_ftp_user"}
)

type Account struct {
	*brightbox.Account
	Role string
}

type AccountsCommand struct {
	*CliApp
	Id     string
	Fields string
}

func AccountFields(a Account) map[string]string {
	return map[string]string{
		"id":                   a.Id,
		"status":               a.Status,
		"name":                 a.Name,
		"role":                 a.Role,
		"cloud_ips_limit":      formatInt(a.CloudIpsLimit),
		"cloud_ips_used":       formatInt(a.CloudIpsUsed),
		"ram_limit":            formatInt(a.RamLimit),
		"ram_used":             formatInt(a.RamUsed),
		"lb_limit":             formatInt(a.LoadBalancersLimit),
		"lb_used":              formatInt(a.LoadBalancersUsed),
		"dbs_ram_limit":        formatInt(a.DbsRamLimit),
		"dbs_ram_used":         formatInt(a.DbsRamUsed),
		"ram_free":             formatInt(a.RamLimit - a.RamUsed),
		"library_ftp_host":     a.LibraryFtpHost,
		"library_ftp_user":     a.LibraryFtpUser,
		"library_ftp_password": a.LibraryFtpPassword,
		"owner_id":             a.Owner.Id,
		"owner_email":          a.Owner.EmailAddress,
		"collaborators":        formatInt(len(a.Users)),
	}
}

func (l *AccountsCommand) list(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	out := new(RowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))

	out.SendHeader()

	accounts, err := l.Client.Accounts()
	if err != nil {
		return err
	}

	colmap := make(map[string]string)
	collabs, err := l.Client.Collaborations()
	// An error here most likely just means that we're using a api client and
	// not a user client. So if it succeeds, we're a user with
	// collaborations. If it fails, we're an api client and don't have
	// collaborations.
	if err == nil {
		for _, a := range *accounts {
			colmap[a.Id] = "Owner"
		}
		for _, col := range *collabs {
			colmap[col.Account.Id] = col.RoleLabel
		}
	}

	for _, a := range *accounts {
		if err = out.Write(AccountFields(Account{&a, colmap[a.Id]})); err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func (l *AccountsCommand) show(pc *kingpin.ParseContext) error {
	err := l.Configure()
	if err != nil {
		return err
	}
	out := new(ShowFieldOutput)
	out.Setup(strings.Split(l.Fields, ","))

	account, err := l.Client.Account(l.Id)
	if err != nil {
		return err
	}
	collabs, err := l.Client.Collaborations()
	colmap := make(map[string]string)
	if err == nil {
		colmap[account.Id] = "Owner"
		for _, col := range *collabs {
			colmap[col.Account.Id] = col.RoleLabel
		}
	}

	if err = out.Write(AccountFields(Account{account, colmap[account.Id]})); err != nil {
		return err
	}
	out.Flush()
	return nil
}

func ConfigureAccountsCommand(app *CliApp) {
	cmd := AccountsCommand{CliApp: app}
	accounts := app.Command("accounts", "manage accounts")

	list := accounts.Command("list", "list accounts").Default().Action(cmd.list)
	list.Flag("fields", "Which fields to display").
		Default(strings.Join(DefaultAccountListFields, ",")).
		StringVar(&cmd.Fields)

	show := accounts.Command("show", "Show detailed account info").Action(cmd.show)
	show.Arg("identifier", "Identifier of account to show").
		Required().StringVar(&cmd.Id)

	show.Flag("fields", "Which fields to display").
		Default(strings.Join(DefaultAccountShowFields, ",")).
		StringVar(&cmd.Fields)

}
