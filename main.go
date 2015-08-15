package main

import (
	"os"
	"./cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cliapp := cli.New()
	kingpin.MustParse(cliapp.Parse(os.Args[1:]))
}
