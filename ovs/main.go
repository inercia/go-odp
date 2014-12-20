package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/dpw/go-openvswitch/openvswitch"
)

func printErr(f string, a ...interface{}) bool {
	fmt.Fprintf(os.Stderr, f, a...)
	os.Stderr.WriteString("\n")
	return false
}

type commandDispatch interface {
	run(args []string, pos int) bool
}

type command func (args []string) bool

func (cmd command) run(args []string, pos int) bool {
	return cmd(args[pos:])
}

type subcommands map[string]commandDispatch

func (cmds subcommands) run(args []string, pos int) bool {
	if pos >= len(args) {
		return printErr("Subcommand required by \"%s\".  Try \"%s help\"", strings.Join(args[:pos], " "), args[0])
	}

	cd, ok := cmds[args[pos]]

	if !ok {
		return printErr("Unknown command \"%s\".  Try \"%s help\"", strings.Join(args[:pos + 1], " "), args[0])
	}

	return cd.run(args, pos + 1)
}

type possibleSubcommands struct {
	command command
	subcommands subcommands
}

func (cmds possibleSubcommands) run(args []string, pos int) bool {
	if pos >= len(args) {
		return cmds.command(args[pos:])
	}

	return cmds.subcommands.run(args, pos)
}


var commands = subcommands {
	"datapath": possibleSubcommands {
		listDatapaths,
		subcommands {
			"create": command(createDatapath),
			"delete": command(deleteDatapath),
		},
	},
}

func main() {
	if (!commands.run(os.Args, 1)) {
		os.Exit(1)
	}
}

func createDatapath(args []string) bool {
	dpif, err := openvswitch.NewDpif()
	if err != nil { return printErr("%s", err) }
	defer dpif.Close()

	for _, name := range(args) {
		_, err = dpif.CreateDatapath(name)
		if err != nil { return printErr("%s", err) }
	}

	return true
}

func deleteDatapath(args []string) bool {
	dpif, err := openvswitch.NewDpif()
	if err != nil { return printErr("%s", err) }
	defer dpif.Close()

	for _, name := range(args) {
		dp, err := dpif.LookupDatapath(name)
		if err != nil { return printErr("%s", err) }

		if dp == nil {
			return printErr("Cannot find datapath \"%s\"", name);
		}

		err = dp.Delete()
		if err != nil { return printErr("%s", err) }
	}

	return true
}

func listDatapaths(args []string) bool {
	dpif, err := openvswitch.NewDpif()
	if err != nil { return printErr("%s", err) }
	defer dpif.Close()

	name2dp, err := dpif.EnumerateDatapaths()
	for name := range(name2dp) {
		fmt.Printf("%s\n", name)
	}

	return true
}
