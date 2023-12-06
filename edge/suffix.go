package edge

import "strings"

type EdgeSuffix struct {
	Production string
	Staging    string
}

func (s EdgeSuffix) IsProd(fqdn string) bool {
	return strings.HasSuffix(fqdn, s.Production)
}

func (s EdgeSuffix) IsStaging(fqdn string) bool {
	return strings.HasSuffix(fqdn, s.Staging)
}

func (s EdgeSuffix) Basify(fqdn string) string {

	if strings.HasSuffix(fqdn, s.Production) {
		return strings.TrimSuffix(fqdn, s.Production)
	}

	if strings.HasSuffix(fqdn, s.Staging) {
		return strings.TrimSuffix(fqdn, s.Staging)
	}

	return fqdn

}

func (s EdgeSuffix) Prodify(fqdn string) string {
	if strings.HasSuffix(fqdn, s.Production) {
		return fqdn
	}
	return strings.TrimSuffix(fqdn, s.Staging) + s.Production
}

func (s EdgeSuffix) Stagify(fqdn string) string {
	return strings.TrimSuffix(fqdn, s.Production) + s.Staging
}

var SUFFIXES = [3]EdgeSuffix{
	{Production: "edgesuite.net.", Staging: "edgesuite-staging.net."},
	{Production: "edgekey.net.", Staging: "edgekey-staging.net."},
	{Production: "akamaiedge.net.", Staging: "akamaiedge-staging.net."},
}
