package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"./cli"
)


func main() {
	app := kingpin.New("brightbox", "Bleh")
	cli.ConfigureServersCommand(app)
	//kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.1").Author("John Leach")
	kingpin.MustParse(app.Parse(os.Args[1:]))
	//kingpin.Parse()
	//fmt.Printf("%v, %s\n", *verbose, *name)
}
