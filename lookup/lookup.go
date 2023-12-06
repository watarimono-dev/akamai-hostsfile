package lookup

import (
	"akamai-hostsfile/edge"

	. "akamai-hostsfile/logging"
	. "akamai-hostsfile/target"

	"github.com/miekg/dns"
)

type LookupResult struct {
	Input        TargetInput
	IsEdgeHosted bool
	Errors       []error
	EdgeHostName edge.EdgeHostName
	Addresses    []string // no point in making the distinction between V4 and V6 addresses ...
}

type LookupConf struct {
	NameServer string
	V4         bool
	V6         bool
}

func (c LookupConf) Qtypes() []uint16 {

	var qtypes []uint16

	if c.V4 {
		qtypes = append(qtypes, dns.TypeA)
	}

	if c.V6 {
		qtypes = append(qtypes, dns.TypeAAAA)
	}

	return qtypes

}

func Lookup(conf LookupConf, target TargetInput) LookupResult {

	Logger.Debug().Str(logFieldTarget, target.FQDN).Msg("starting lookup")

	result := LookupResult{Input: target}

	q := query(conf.NameServer, dns.TypeCNAME, target.FQDN)
	if q.Error != nil {
		result.Errors = append(result.Errors, q.Error)
		return result
	}

	for _, cname := range q.Results {
		if edgeHostName, err := edge.NewEdgeHostNameFromFQDN(cname); err == nil {
			result.IsEdgeHosted = true
			result.EdgeHostName = edgeHostName
			break
		}
	}

	// no point in resolving target if it is not edge hosted
	if !result.IsEdgeHosted {
		Logger.Debug().Str(logFieldTarget, target.FQDN).Str(logFieldTarget, target.FQDN).Msg("interrupting lookup because target does not seem to be edge hosted")
		return result
	}

	qtypes := conf.Qtypes()
	qchan := make(chan QueryResult, len(qtypes))

	for _, qtype := range qtypes {

		go func(qtype uint16) {
			qchan <- query(conf.NameServer, qtype, result.EdgeHostName.Network(target.TargetNetwork))
		}(qtype)

	}

	for i := 0; i < len(qtypes); i++ {
		r := <-qchan
		if r.Error != nil {
			result.Errors = append(result.Errors, r.Error)
		} else {
			result.Addresses = append(result.Addresses, r.Results...)
		}
	}

	close(qchan)

	return result

}

func LookupMany(conf LookupConf, sets ...map[string]TargetInput) map[int][]LookupResult {

	indexedSets := make(map[int]map[string]TargetInput)
	indexedChan := make(map[int]chan LookupResult)
	indexedResults := make(map[int][]LookupResult)

	for i, set := range sets {
		indexedSets[i] = set
		indexedChan[i] = make(chan LookupResult, len(set))
	}

	// iterate over sets because iterating over maps might not respect the order
	for i := range sets {

		for _, input := range indexedSets[i] {

			go func(i int, input TargetInput) {
				indexedChan[i] <- Lookup(conf, input)
			}(i, input)

		}

	}

	// iterate over sets because iterating over maps might not respect the order
	for i := range sets {

		for x := 0; x < len(indexedSets[i]); x++ {
			indexedResults[i] = append(indexedResults[i], <-indexedChan[i])
		}

	}

	for i := range sets {
		close(indexedChan[i])
	}

	return indexedResults

}
