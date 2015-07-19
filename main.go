package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/miekg/dns"
	"regexp"
)

//TODO MORE PROTOCOLS

type domain struct {
	Host, Domain, Password, Protocol, ip, fqdn string
	getIP                                      chan string
}

var validIP = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

func (c *domain) listenIPLoop() {
	if c.Host == "@" {
		c.fqdn = c.Domain + "."
	} else {
		c.fqdn = c.Host + "." + c.Domain + "."
	}
	var dnsIP string
	dmsg := new(dns.Msg)
	dmsg.SetQuestion(c.fqdn, dns.TypeA)
	for newIP := range c.getIP {
		log.Println("checking ip for", c.fqdn)
		in, err := dns.Exchange(dmsg, "dns1.registrar-servers.com:53")
		if err != nil {
			log.Println(err)
			continue
		}
		if len(in.Answer) >= 1 {
			dnsIP = validIP.FindString(in.Answer[0].String())
			log.Printf("current ip for %s is %s", c.fqdn, dnsIP)
		} else {
			log.Printf("could not find A record for %s", c.fqdn)
			continue
		}
		if newIP != dnsIP {
			err := c.updateIP(newIP)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(c.fqdn, "up to date")
		}
	}
}

func (c *domain) useProtocol() (r *http.Response, err error) {
	switch strings.ToLower(c.Protocol) {
	case "namecheap":
		r, err = http.Get("https://dynamicdns.park-your-domain.com/update?host=" + c.Host + "&domain=" + c.Domain + "&password=" + c.Password + "&ip=" + c.ip)
	}
	return
}

func (c *domain) checkError(buf []byte, r *http.Response) (errStr string, bad bool) {
	s := (string(buf))
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
		errStr = fmt.Sprintf("status code %d; could not update %s; %s", r.StatusCode, c.fqdn, s)
	}
	return
}

func (c *domain) updateIP(newIP string) error {
	log.Println("updating", c.fqdn)
	r, err := c.useProtocol()
	if err != nil {
		return err
	}
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if errStr, bad := c.checkError(buf, r); bad {
		err := errors.New(errStr)
		return err
	}
	log.Println("updated", c.fqdn)
	return nil
}

func (config *configuration) parseConfig() {
	log.Println("reading config.json")
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read config.json")
	log.Println("parsing config.json")
	d := json.NewDecoder(f)
	err = d.Decode(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("parsed config.json")
}

func (config *configuration)checkIPLoop() {
	for ;; time.Sleep(time.Second * config.Interval) {
		log.Println("getting public IP")
		resp, err := http.Get("http://echoip.com")
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		newIP := string(buf)
		log.Println("got", newIP)
		log.Println("sending IP to goroutines")
		for _, c := range config.List {
			c.getIP <- newIP
		}
		log.Println("sent IP to goroutines")
	}
}

type configuration struct {
	List     []*domain
	Interval time.Duration
}

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

//example config
//{
//"List": [
//{
//"Host": "@",
//"Domain": "aubble.com",
//"Password": "PASSWORDHERE",
//"Protocol": "namecheap"
//}
//],
//"Interval": 20 //seconds
//} //REMOVE ALL COMMENTS PLEASE
