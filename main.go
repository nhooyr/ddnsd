package main

import (
	"encoding/json"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//TODO MORE PROTOCOLS

func main() {
	log.SetPrefix("goDDNS: ")
	config := configuration{}
	config.parseConfig()
	log.Println("launching goroutines")
	for _, c := range config.List {
		c.getIP = make(chan string)
		go c.listenIPLoop()
	}
	log.Println("launched goroutines")
	config.checkIPLoop()
}

type configuration struct {
	List     []*domain
	Interval time.Duration
}

func (c *configuration) parseConfig() {
	log.Println("reading config.json")
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read config.json")
	log.Println("parsing config.json")
	d := json.NewDecoder(f)
	err = d.Decode(c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("parsed config.json")
}

func (c *configuration) checkIPLoop() {
	for ;; time.Sleep(time.Second * c.Interval) {
		newIP, err := c.getPublicIP()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("got", newIP)
		log.Println("sending IP to goroutines")
		for _, c := range c.List {
			c.getIP <- newIP
		}
		log.Println("sent IP to goroutines")
	}
}

func (c *configuration) getPublicIP() (string, error) {
	log.Println("getting public IP")
	resp, err := http.Get("http://echoip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

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
	return
}
