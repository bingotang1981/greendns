package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const VERSION = "1.0.0"

var (
	ruleMap        map[string]string
	filterFilePath string
	port           string = "10053"
	upstreamDns    string = "114.114.114.114:53"
	defaultFakeIp  string = "8.8.8.255"
)

type dnsHandler struct{}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// fmt.Println(r.Question[0].Name, r.Question[0].Qtype)

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = false

	for _, question := range r.Question {
		answers := resolver(question.Name, question.Qtype, r)
		msg.Answer = append(msg.Answer, answers...)
	}
	w.WriteMsg(msg)
}

func resolver(domain string, qtype uint16, r *dns.Msg) []dns.RR {
	//skip Type AAAA (ipv6)
	if qtype == dns.TypeAAAA {
		return nil
	}

	if FastMatchDomain(strings.TrimRight(domain, ".")) {
		//Only reply TypeA question
		if qtype == dns.TypeA {
			a := new(dns.A)
			a.Hdr.Name = domain
			a.Hdr.Rrtype = dns.TypeA
			a.Hdr.Ttl = 600
			a.A = net.ParseIP(defaultFakeIp)
			a.Hdr.Class = dns.ClassINET
			m := new(dns.Msg)
			m.SetReply(r)
			m.Answer = []dns.RR{a}
			// for _, answer := range m.Answer {
			// 	fmt.Printf("%s\n", answer.String())
			// }
			return m.Answer
		} else {
			return nil
		}
	} else {

		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), qtype)
		m.RecursionDesired = true

		c := &dns.Client{Timeout: 5 * time.Second}

		response, _, err := c.Exchange(m, upstreamDns)
		if err != nil {
			// fmt.Printf("[ERROR] : %v\n", err)
			return nil
		}

		if response == nil {
			// fmt.Printf("[ERROR] : no response from server\n")
			return nil
		}

		// for _, answer := range response.Answer {
		// 	fmt.Printf("%s\n", answer.String())
		// }

		return response.Answer
	}
}

//This method parses the filter file
func ParseFilterFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	ruleMap = make(map[string]string)
	for {
		line, _, err := reader.ReadLine()
		if nil != err {
			break
		}
		str := strings.TrimSpace(string(line))
		if str != "" {
			ruleMap[str] = ""
		}
	}
	return nil
}

//Returns whether it is in the filter map
func FastMatchDomain(domain string) bool {
	if ruleMap == nil {
		return false
	}

	myDomain := domain

	ds := strings.Split(myDomain, ".")

	num := len(ds)
	if num == 1 {
		return false
	} else if num == 2 {
		dm := myDomain
		_, flag := ruleMap[dm]
		return flag
	} else if num == 3 {
		dm := myDomain
		_, flag := ruleMap[dm]
		if flag {
			return true
		} else {
			dm := ds[num-2] + "." + ds[num-1]
			_, flag := ruleMap[dm]
			return flag
		}
	} else {
		dm := ds[num-4] + "." + ds[num-3] + "." + ds[num-2] + "." + ds[num-1]
		_, flag := ruleMap[dm]
		if flag {
			return true
		} else {
			dm := ds[num-3] + "." + ds[num-2] + "." + ds[num-1]
			_, flag := ruleMap[dm]
			if flag {
				return true
			} else {
				dm := ds[num-2] + "." + ds[num-1]
				_, flag := ruleMap[dm]
				return flag
			}
		}
	}
}

func main() {
	fmt.Println("VERSION:", VERSION)
	fmt.Println("greendns <filterFilePath> [port] [upstreamDns] [defaultFakeIp]")

	arg_num := len(os.Args)

	if arg_num < 2 {
		return
	}

	filterFilePath = os.Args[1]

	if arg_num > 2 {
		port = os.Args[2]
	}

	if arg_num > 3 {
		upstreamDns = os.Args[3]
	}

	if arg_num > 4 {
		defaultFakeIp = os.Args[4]
	}

	err := ParseFilterFile(filterFilePath)

	if err != nil {
		fmt.Println("Fail to load filter file", err)
		return
	}

	handler := new(dnsHandler)
	server := &dns.Server{
		Addr:      ":" + port,
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	fmt.Println("Starting DNS server on port", port)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start server", err)
	}
}
