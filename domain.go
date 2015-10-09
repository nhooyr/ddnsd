package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type domain struct {
	Host, Domain, Password, Protocol, ip, fqdn, dnsIP string
	getIP                                             chan string
	dnsMsg                                            *dns.Msg
	c                                                 *configuration
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
			d.log(d.fqdn, "up to date")
		}
	}
}

var validIP = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

func (d *domain) checkIP() {
	d.log("checking ip for", d.fqdn)
	in, err := dns.Exchange(d.dnsMsg, "dns1.registrar-servers.com:53")
	if err != nil {
		d.log(err)
		return
	}
	if len(in.Answer) >= 1 {
		d.dnsIP = validIP.FindString(in.Answer[0].String())
		d.log("current ip for", d.fqdn, "is", d.dnsIP)
	} else {
		d.log("could not find A record for", d.fqdn)
	}
}

func (d *domain) updateIP(newIP string) {
	d.log("updating", d.fqdn)
	r, err := d.useProtocol()
	if err != nil {
		d.log(err)
		return
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		d.log(err)
		return
	}
	if bad := d.checkError(string(buf), r); bad {
		return
	}
	d.log("updated", d.fqdn)
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
			d.log("status code", strconv.Itoa(r.StatusCode)+"; could not update;", d.fqdn+";", s)
		}
	}
	return
}
func (d *domain) log(v ...interface{}) {
	logger.println(append([]interface{}{"-->", d.Host + "." + d.Domain, "-/"}, v...)...)
}
