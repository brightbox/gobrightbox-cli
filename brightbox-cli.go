package main

import (
	"os"
	"log"
)

func main() {
	var	err os.Error
	var config *Config
	config, err = NewConfig()
	if err != nil {
		log.Fatal("Couldn't read config: ", err)
		os.Exit(1)
	}

	auth := config.Auth()
	err = config.SetupAuthenticatorCache(auth)

	servers := config.Client().ListServers()

	table_data := make([][]string, len(servers))
	for i, s := range servers {
		table_data[i] = []string{s.Id, s.Status, s.ServerType.Handle, s.Zone.Handle, "", s.Image.Id, "", s.Name}
	}
	DrawTable([]string{"id", "status", "type", "zone", "created_on", "image_id", "cloud_ip_ids", "name"}, table_data)
	return
}