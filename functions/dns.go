package functions

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/miekg/dns"
)

var (
	// DefaultDnsTimeout is the default timeout for DNS lookups
	DefaultDnsTimeout = 10 * time.Second
)

// httpsLookup performs HTTPS record lookup using miekg/dns library
func httpsLookup(hostname string) (any, error) {
	c := dns.Client{Timeout: DefaultDnsTimeout}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(hostname), dns.TypeHTTPS)

	r, _, err := c.Exchange(&m, "1.1.1.1:53") // Use Cloudflare DNS
	if err != nil {
		return nil, fmt.Errorf("dns_lookup: HTTPS record lookup failed: %w", err)
	}

	if len(r.Answer) == 0 {
		return map[string]any{
			"hostname": hostname,
			"type":     "HTTPS",
			"success":  true,
			"records":  []any{},
		}, nil
	}

	var records []any
	for _, ans := range r.Answer {
		if https, ok := ans.(*dns.HTTPS); ok {
			record := map[string]any{
				"priority": int(https.Priority),
				"target":   strings.TrimSuffix(https.Target, "."),
				"params":   map[string]any{},
			}

			// Parse service parameters
			for _, param := range https.Value {
				switch param.Key() {
				case dns.SVCB_ALPN:
					if alpn, ok := param.(*dns.SVCBAlpn); ok {
						var alpnList []any
						for _, protocol := range alpn.Alpn {
							alpnList = append(alpnList, protocol)
						}
						record["params"].(map[string]any)["alpn"] = alpnList
					}
				case dns.SVCB_PORT:
					if port, ok := param.(*dns.SVCBPort); ok {
						record["params"].(map[string]any)["port"] = int(port.Port)
					}
				case dns.SVCB_IPV4HINT:
					if ipv4, ok := param.(*dns.SVCBIPv4Hint); ok {
						var ips []any
						for _, ip := range ipv4.Hint {
							ips = append(ips, ip.String())
						}
						record["params"].(map[string]any)["ipv4hint"] = ips
					}
				case dns.SVCB_IPV6HINT:
					if ipv6, ok := param.(*dns.SVCBIPv6Hint); ok {
						var ips []any
						for _, ip := range ipv6.Hint {
							ips = append(ips, ip.String())
						}
						record["params"].(map[string]any)["ipv6hint"] = ips
					}
				}
			}
			records = append(records, record)
		}
	}

	return map[string]any{
		"hostname": hostname,
		"type":     "HTTPS",
		"success":  true,
		"records":  records,
	}, nil
}

// dnslookup performs DNS lookup for the specified hostname and record type
func dnslookup(hostname, recordType string) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDnsTimeout)
	defer cancel()

	resolver := &net.Resolver{}
	recordType = strings.ToUpper(recordType)

	result := map[string]any{
		"hostname": hostname,
		"type":     recordType,
		"success":  true,
	}

	switch recordType {
	case "A":
		ips, err := resolver.LookupIPAddr(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: A record lookup failed: %w", err)
		}
		var records []any
		for _, ip := range ips {
			if ip.IP.To4() != nil { // IPv4 only
				records = append(records, ip.IP.String())
			}
		}
		result["records"] = records

	case "AAAA":
		ips, err := resolver.LookupIPAddr(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: AAAA record lookup failed: %w", err)
		}
		var records []any
		for _, ip := range ips {
			if ip.IP.To4() == nil { // IPv6 only
				records = append(records, ip.IP.String())
			}
		}
		result["records"] = records

	case "MX":
		mxRecords, err := resolver.LookupMX(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: MX record lookup failed: %w", err)
		}
		var records []any
		for _, mx := range mxRecords {
			records = append(records, map[string]any{
				"priority": int(mx.Pref),
				"hostname": strings.TrimSuffix(mx.Host, "."),
			})
		}
		result["records"] = records

	case "TXT":
		txtRecords, err := resolver.LookupTXT(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: TXT record lookup failed: %w", err)
		}
		var records []any
		for _, txt := range txtRecords {
			records = append(records, txt)
		}
		result["records"] = records

	case "PTR":
		names, err := resolver.LookupAddr(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: PTR record lookup failed: %w", err)
		}
		var records []any
		for _, name := range names {
			records = append(records, strings.TrimSuffix(name, "."))
		}
		result["records"] = records

	case "CNAME":
		cname, err := resolver.LookupCNAME(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: CNAME record lookup failed: %w", err)
		}
		result["records"] = []any{strings.TrimSuffix(cname, ".")}

	case "NS":
		nsRecords, err := resolver.LookupNS(ctx, hostname)
		if err != nil {
			return nil, fmt.Errorf("dns_lookup: NS record lookup failed: %w", err)
		}
		var records []any
		for _, ns := range nsRecords {
			records = append(records, strings.TrimSuffix(ns.Host, "."))
		}
		result["records"] = records

	case "HTTPS":
		return httpsLookup(hostname)

	case "SVCB":
		// SVCB records are similar to HTTPS but for other services
		// For now, we'll treat them the same as HTTPS records but with different type
		result, err := httpsLookup(hostname)
		if err != nil {
			return nil, err
		}
		if resultMap, ok := result.(map[string]any); ok {
			resultMap["type"] = "SVCB"
		}
		return result, nil

	default:
		return nil, fmt.Errorf("dns_lookup: unsupported record type: %s (supported: A, AAAA, MX, TXT, PTR, CNAME, NS, HTTPS, SVCB)", recordType)
	}

	return result, nil
}

var DnsFunctions = map[string]*jsonnet.NativeFunction{
	"dns_lookup": {
		Params: []ast.Identifier{"hostname", "record_type"},
		Func: func(args []any) (any, error) {
			hostname, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("dns_lookup: hostname must be a string")
			}

			recordType, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("dns_lookup: record_type must be a string")
			}

			return dnslookup(hostname, recordType)
		},
	},
}

func init() {
	initializeFunctionMap(DnsFunctions)
}
