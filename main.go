package main

import (
	"akamai-hostsfile/lookup"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {

	mainCommand.Flags().String("edgerc", "", "not used but implemented because it is passed by default by the akamai CLI")
	mainCommand.Flags().String("section", "", "not used but implemented because it is passed by default by the akamai CLI")

	mainCommand.Flags().StringP("hostsfile", "f", "/etc/hosts", "path to the hostsfile")
	mainCommand.Flags().StringP("nameserver", "n", "8.8.8.8:53", "the nameserver server to send queries to (formatted as \"{host}:{port})\" - the port defaults to 53 if none is provided")

	mainCommand.Flags().BoolP("clean", "c", false, "deletes Edge hosted entries from the hostsfile")
	mainCommand.Flags().BoolP("infile", "i", false, "write changes to the hostsfile instead of simply showing the final output")
	mainCommand.Flags().BoolP("v4", "4", false, "resolve IPv4 target addresses (defaults to true unless only --v6/-6 is passed)")
	mainCommand.Flags().BoolP("v6", "6", false, "resolve IPv6 target addresses (defaults to true unless only --v4/-4 is passed))")

	mainCommand.Flags().StringArrayP("prod", "p", nil, "write entry pointing to the production network IP addresses for the specified target")
	mainCommand.Flags().StringArrayP("staging", "s", nil, "write entry pointing to the staging network IP addresses for the specified target")

}

func main() {

	mainCommand.Execute()

}

var mainCommand = cobra.Command{

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New(fmt.Sprintf("the command does not take any positional arguments but got : %v", args))
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {

		session := NewSessionFromArgs(cmd)

		results := lookup.LookupMany(session.LookupConf, session.Targets.Staging, session.Targets.Prod, session.Targets.Hostsfile)
		argResults := append(results[0], results[1]...)
		fileResults := results[2]

		if session.Flags.Clean {
			session.Clean(fileResults...)
		}

		for _, result := range argResults {
			session.Clean(result)
			session.Write(result)
		}

		session.Close()

	},
}
