package edge

import (
	"errors"
	"fmt"
)

const (
	PRODUCTION = iota
	STAGING
)

type EdgeHostName struct {
	Production string
	Staging    string
	Base       string
	Suffix     EdgeSuffix
}

func NewEdgeHostNameFromFQDN(fqdn string) (EdgeHostName, error) {

	// In the theoric event where we'd have a considerable number of suffixes, iterating over
	// an array of EdgeSuffix would not be the most efficient way. Instead, we could organise
	// the suffixes in a map, and retrieve the EdgeSuffix by parsing the fqdn.
	// This theoric event seems unlikely though, so for the sake of keeping things simple, we'll
	// just iterate over an array

	if fqdn[len(fqdn)-1] != '.' {
		return EdgeHostName{}, errors.New("target is not a FQDN")
	}

	for _, suffix := range SUFFIXES {

		if suffix.IsProd(fqdn) {
			return EdgeHostName{
				Production: fqdn,
				Staging:    suffix.Stagify(fqdn),
				Base:       suffix.Basify(fqdn),
				Suffix:     suffix,
			}, nil
		}

		if suffix.IsStaging(fqdn) {
			return EdgeHostName{
				Production: suffix.Prodify(fqdn),
				Staging:    fqdn,
				Base:       suffix.Basify(fqdn),
				Suffix:     suffix,
			}, nil
		}

	}

	return EdgeHostName{}, errors.New("hostname did not match any known edge hostname suffix")

}

func (e EdgeHostName) Network(network int) string {

	if network == PRODUCTION {
		return e.Production
	}

	if network == STAGING {
		return e.Staging
	}

	panic(fmt.Sprintf("invalid network %d for EdgeHostName %s", network, e))

}
