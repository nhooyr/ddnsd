package main

import (
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type domain struct {
	Host, Domain, Password, Protocol, ip, fqdn, dnsIP string
	getIP                                             chan string
	dnsMsg                                            *dns.Msg
}

func (d *domain) listenIPLoop() {
	if d.Host == "@" {
		d.fqdn = d.Domain + "."
	} else {
		d.fqdn = d.Host + "." + d.Domain + "."
	}
	d.dnsMsg = new(dns.Msg)
	d.dnsMsg.SetQuestion(d.fqdn, dns.TypeA)
	for newIP := range d.getIP {
		d.checkIP()
		if newIP != d.dnsIP {
			d.updateIP(newIP)
		} else {
			log.Println(d.fqdn, "up to date")
		}
	}
}

var validIP = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

func (d *domain) checkIP() {
	log.Println("checking ip for", d.fqdn)
	in, err := dns.Exchange(d.dnsMsg, "dns1.registrar-servers.com:53")
	if err != nil {
		log.Println(err)
		return
	}
	if len(in.Answer) >= 1 {
		d.dnsIP = validIP.FindString(in.Answer[0].String())
		log.Printf("current ip for %s is %s", d.fqdn, d.dnsIP)
	} else {
		log.Println("could not find A record for", d.fqdn)
	}
}

func (d *domain) updateIP(newIP string) {
	log.Println("updating", d.fqdn)
	r, err := d.useProtocol()
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	if bad := d.checkError(string(buf), r); bad {
		return
	}
	log.Println("updated", d.fqdn)
}

func (d *domain) useProtocol() (r *http.Response, err error) {
	switch strings.ToLower(d.Protocol) {
	case "namecheap":
		r, err = http.Get("https://dynamicdns.park-your-domain.com/update?host=" + d.Host + "&domain=" + d.Domain + "&password=" + d.Password + "&ip=" + d.ip)
	}
	return
}

func (d *domain) checkError(s string, r *http.Response) (bad bool) {
	switch strings.ToLower(d.Protocol) {
	case "namecheap":
		if strings.Contains(strings.ToLower(s), "error") {
			bad = true
			i := strings.Index(s, "<ResponseString>")
			if i != -1 {
				j := strings.Index(s, "</ResponseString>")
				s = s[i+len("<ResponseString>") : j]
			} else {
				i = strings.Index(s, "<p>")
				j := strings.Index(s, "</p>")
				s = s[i+len("<p>") : j]
			}
			log.Printf("status code %d; could not update %s; %s", r.StatusCode, d.fqdn, s)
		}
	}
	return
}
