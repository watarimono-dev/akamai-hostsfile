package main

import (
	"akamai-hostsfile/edge"
	"akamai-hostsfile/lookup"
	"akamai-hostsfile/target"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	. "akamai-hostsfile/logging"

	"github.com/goodhosts/hostsfile"
	"github.com/spf13/cobra"
)

var logFieldHostsFilePath = "hostsfile"
var logFieldFlagInfile = "infile"
var logFieldAddr = "addr"
var logFieldLineNo = "lineno"
var logFieldTarget = "target"
var logFieldNameServer = "nameserver"

type Targets struct {
	Staging   map[string]target.TargetInput
	Prod      map[string]target.TargetInput
	Hostsfile map[string]target.TargetInput
}

type ProcessFlags struct {
	Clean  bool
	Infile bool
}

type Session struct {
	Hostsfile  *hostsfile.Hosts
	Targets    Targets
	LookupConf lookup.LookupConf
	Flags      ProcessFlags
}

func NewSessionFromArgs(cmd *cobra.Command) Session {

	hostsfilepath, _ := cmd.Flags().GetString("hostsfile")
	nameserver, _ := cmd.Flags().GetString("nameserver")

	clean, _ := cmd.Flags().GetBool("clean")
	infile, _ := cmd.Flags().GetBool("infile")
	v4, _ := cmd.Flags().GetBool("v4")
	v6, _ := cmd.Flags().GetBool("v6")

	staging, _ := cmd.Flags().GetStringArray("staging")
	prod, _ := cmd.Flags().GetStringArray("prod")

	lookupConf := lookup.LookupConf{
		NameServer: InitNameServer(nameserver),
		V4:         v4,
		V6:         v6,
	}
	if v4 == v6 {
		lookupConf.V4 = true
		lookupConf.V6 = true
	}

	flags := ProcessFlags{
		Clean:  clean,
		Infile: infile,
	}

	file, _ := InitHostsFile(hostsfilepath, flags)
	targets := InitTargets(staging, prod, file)

	return Session{
		Hostsfile:  file,
		Targets:    targets,
		LookupConf: lookupConf,
		Flags:      flags,
	}

}

func InitNameServer(nameserver string) string {

	ns := strings.Split(nameserver, ":")

	switch len(ns) {

	case 1:
		return nameserver + ":53"

	case 2:
		if _, err := strconv.Atoi(ns[1]); err == nil {
			return nameserver
		}
		fallthrough

	default:
		Logger.Fatal().Str(logFieldNameServer, nameserver).Msg("malformed nameserver argument")
		panic("unreachable") // missing return
	}

}

func InitHostsFile(filepath string, flags ProcessFlags) (*hostsfile.Hosts, error) {

	file, err := hostsfile.NewCustomHosts(filepath)

	if err != nil {
		Logger.Error().Err(err).Send()
		return file, err
	}

	if !file.IsWritable() {
		if flags.Infile {
			Logger.Fatal().Str(logFieldHostsFilePath, filepath).Msg("hostsfile is not writable but --infile/-i was passed")
			// Logger.Fatal exists the program
		}
		Logger.Debug().Str(logFieldHostsFilePath, filepath).Msg("hostsfile is writable")
	}

	return file, nil

}

func InitTargets(staging []string, prod []string, hostsfile *hostsfile.Hosts) Targets {

	// The TargetInput.FQDN is used as key as it uniquely identifies a target, where TargetInput.Input could take both FQDN and non FQDN forms resulting in duplicates
	targets := Targets{
		Staging:   make(map[string]target.TargetInput),
		Prod:      make(map[string]target.TargetInput),
		Hostsfile: make(map[string]target.TargetInput),
	}

	for _, input := range staging {
		target := target.New(edge.STAGING, input)
		targets.Staging[target.FQDN] = target
	}

	for _, input := range prod {
		target := target.New(edge.PRODUCTION, input)
		targets.Prod[target.FQDN] = target
	}

	for lineno, line := range hostsfile.Lines {

		if len(line.IP) == 0 {
			continue
		}

		addr, err := netip.ParseAddr(line.IP)

		if err != nil {
			Logger.Error().Err(err).Str(logFieldAddr, line.IP).Str(logFieldAddr, line.IP).Int(logFieldLineNo, lineno).Msg("error parsing address")
			continue
		}

		if addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() || addr.IsMulticast() || line.IP == "fe00::0" {
			Logger.Debug().Str(logFieldAddr, line.IP).Msg("skipping")
			continue
		}

		for _, host := range line.Hosts {

			target := target.New(edge.PRODUCTION, host) // The network is not relevant here and could be either one of them

			if _, ok := targets.Prod[target.FQDN]; ok {
				continue
			}
			if _, ok := targets.Staging[target.FQDN]; ok {
				continue
			}

			targets.Hostsfile[target.FQDN] = target

		}

	}

	return targets

}

func (s Session) Clean(results ...lookup.LookupResult) {

	for _, result := range results {

		if result.IsEdgeHosted {

			Logger.Debug().Str(logFieldTarget, result.Input.FQDN).Msg("cleaning entries")

			s.Hostsfile.RemoveByHostname(result.Input.Raw)
			s.Hostsfile.RemoveByHostname(result.Input.FQDN)

		}

	}

}

func (s Session) Write(result lookup.LookupResult) {

	for _, addr := range result.Addresses {

		Logger.Debug().Str(logFieldTarget, result.EdgeHostName.Base).Str(logFieldAddr, addr).Msg("adding entry")

		newLine := hostsfile.HostsLine{IP: addr, Hosts: []string{result.EdgeHostName.Base}, Comment: "---> " + result.EdgeHostName.Network(result.Input.TargetNetwork)}
		s.Hostsfile.Lines = append(s.Hostsfile.Lines, newLine)

	}

}

func (s Session) Close() {

	Logger.Debug().Bool(logFieldFlagInfile, s.Flags.Infile).Msg("closing session")

	if s.Flags.Infile {
		s.Hostsfile.Flush()
	}

	fmt.Println(s.Hostsfile)

}
