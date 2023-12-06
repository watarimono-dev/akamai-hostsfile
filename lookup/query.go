package lookup

import (
	"errors"
	"fmt"

	. "akamai-hostsfile/logging"

	"github.com/miekg/dns"
)

var logFieldTarget = "target"
var logFieldNameServer = "nameserver"
var logFieldQtype = "qtype"
var logFieldRecords = "records"

func qtypeName(qtype uint16) string {

	switch qtype {

	case dns.TypeCNAME:
		return "CNAME"

	case dns.TypeA:
		return "A"

	case dns.TypeAAAA:
		return "AAAA"

	default:
		Logger.Error().Err(errors.New(fmt.Sprintf("qtypeName not implemented for qtype=%d", qtype))).Send()
		return "<ERR>"
	}

}

type QueryResult struct {
	Results []string
	Error   error
}

func query(nameserver string, qtype uint16, target string) QueryResult {

	result := QueryResult{}

	q := new(dns.Msg)
	q.SetQuestion(target, qtype)

	a, err := dns.Exchange(q, nameserver)
	if err != nil {
		Logger.Error().Err(err).Str(logFieldTarget, target).Str(logFieldNameServer, nameserver).Str(logFieldQtype, qtypeName(qtype)).Send()
		result.Error = err
		return result
	}

	for _, rr := range a.Answer {

		if rr.Header().Rrtype == qtype {
			switch t := rr.(type) {
			case *dns.A:
				result.Results = append(result.Results, t.A.String())
			case *dns.AAAA:
				result.Results = append(result.Results, t.AAAA.String())
			case *dns.CNAME:
				result.Results = append(result.Results, t.Target)
			}
		}

	}

	if len(result.Results) == 0 {
		result.Error = errors.New(fmt.Sprintf("empty answer for query nameserver=%s qtype=%s target=%s", nameserver, qtypeName(qtype), target))
		Logger.Error().Err(result.Error).Str(logFieldTarget, target).Str(logFieldNameServer, nameserver).Str(logFieldQtype, qtypeName(qtype)).Send()
		return result
	}

	Logger.Info().Str(logFieldTarget, target).Str(logFieldNameServer, nameserver).Str(logFieldQtype, qtypeName(qtype)).Any(logFieldRecords, result.Results).Send()
	return result

}
