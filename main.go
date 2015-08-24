package main

import (
	"github.com/brightbox/gobrightbox-cli/cli"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func main() {
	cliapp := cli.New()
	kingpin.MustParse(cliapp.Parse(os.Args[1:]))
}
