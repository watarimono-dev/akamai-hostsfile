package target

import (
	"github.com/miekg/dns"
)

const (
	HOSTSFILE = iota
	ARG
)

type TargetInput struct {
	Raw           string
	FQDN          string
	TargetNetwork int
}

func New(network int, target string) TargetInput {
	return TargetInput{
		Raw:           target,
		FQDN:          dns.Fqdn(target),
		TargetNetwork: network,
	}
}
