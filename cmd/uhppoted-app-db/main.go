package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/command"
	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-app-db/commands"
)

var cli = []uhppoted.Command{
	&commands.LoadACLCmd,
	&commands.StoreACLCmd,
	&commands.GetACLCmd,
	&commands.PutACLCmd,

	&uhppoted.Version{
		Application: commands.APP,
		Version:     uhppote.VERSION,
	},
}

var help = uhppoted.NewHelp(commands.APP, cli, nil)

var options = commands.Options{
	Config: config.DefaultConfig,
	Debug:  false,
}

func main() {
	flag.StringVar(&options.Config, "config", options.Config, "configuration file to use for controller identification and configuration")
	flag.BoolVar(&options.Debug, "debug", options.Debug, "Enable debugging information")
	flag.Parse()

	cmd, err := uhppoted.Parse(cli, nil, help)
	if err != nil {
		fmt.Printf("\nError parsing command line: %v\n\n", err)
		os.Exit(1)
	}

	if cmd == nil {
		help.Execute()
		os.Exit(1)
	}

	if err = cmd.Execute(&options); err != nil {
		fmt.Printf("\n   ERROR: %v\n\n", err)
		os.Exit(1)
	}
}
