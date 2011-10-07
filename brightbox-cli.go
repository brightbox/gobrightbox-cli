package main

import (
	"brightbox"
	"tabwriter"
	"fmt"
	"os"
	"log"
	config "github.com/kless/goconfig/config"
)

func main() {
	var (
		conf  *config.Config
		err   os.Error
	)
	conf_filename := os.Getenv("HOME") + "/.brightbox/config"
	conf, err = config.ReadDefault(conf_filename)
	if err != nil {
		log.Fatal("Error reading config file: ", conf_filename)
	}
	client_name := os.Getenv("CLIENT")
	if client_name == "" {
		log.Fatal("You must specify the config section name in environment variable CLIENT")
	}
	auth_url, _ := conf.String(client_name, "auth_url")
	api_url, _ := conf.String(client_name, "api_url")
	client_id, _ := conf.String(client_name, "client_id")
	secret, _ := conf.String(client_name, "secret")
	if auth_url == "" && api_url != "" {
		auth_url = api_url
	}

	auth := brightbox.NewApiClientAuth(auth_url, client_id, secret)

	brightbox.SetupAuthenticatorCache(auth)
	client := brightbox.NewClient(api_url, "1.0", auth)
	servers := client.ListServers()

	out := tabwriter.NewWriter(os.Stdout, 4, 2, 2, ' ', tabwriter.StripEscape)
	out.Write([]byte(fmt.Sprintf(" id\tstatus\ttype\tzone\tcreated_on\timage_id\tcloud_ip_ids\tname\n")))
	out.Write([]byte("\t\t\t\t\t\t\t\n"))
	for _, server := range servers {
		out.Write([]byte(fmt.Sprintf(" %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", server.Id, server.Status, server.ServerType.Handle, server.Zone.Handle, "", "", "", server.Name)))
	}
	out.Flush()

	return
}