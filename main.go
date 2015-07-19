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

//TODO MORE PROTOCOLS, USE CHECKDNS VERSION

type config struct {
	Host, Domain, Password, Protocol, ip, fqdn string
	getIP                                      chan string
}

func (c *config) listenIPLoop() {
	if c.Host == "@" {
		c.fqdn = c.Domain + "."
	} else {
		c.fqdn = c.Host + "." + c.Domain + "."
	}
	var dnsIP string
	dmsg := new(dns.Msg)
	dmsg.SetQuestion(c.fqdn, dns.TypeA)
	validIP := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	for newIP := range c.getIP {
		if newIP != dnsIP {
			err := c.updateIP(newIP)
			if err != nil {
				log.Println(err)
			}
			in, err := dns.Exchange(dmsg, "dns1.registrar-servers.com:53")
			if err != nil {
				log.Println(err)
				continue
			}
			if len(in.Answer) >= 1 {
				dnsIP = validIP.FindString(in.Answer[0].String())
			} else {
				log.Println("cannot set", c.fqdn)
			}
		} else {
			log.Println(c.fqdn, "up to date")
		}
	}
}

func (c *config) useProtocol() (r *http.Response, err error) {
	switch strings.ToLower(c.Protocol) {
	case "namecheap":
		r, err = http.Get("https://dynamicdns.park-your-domain.com/update?host=" + c.Host + "&domain=" + c.Domain + "&password=" + c.Password + "&ip=" + c.ip)
	}
	return
}

func (c *config) checkError(buf []byte, r *http.Response) (errStr string, bad bool) {
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
		errStr = fmt.Sprintf("could not update status code %d; %s %s", r.StatusCode, c.fqdn, s)
	}
	return
}

func (c *config) updateIP(newIP string) error {
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

func parseConfig() (config configuration) {
	log.Println("reading config.json")
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read config.json")
	log.Println("parsing config.json")
	config = configuration{}
	d := json.NewDecoder(f)
	err = d.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("parsed config.json")
	return
}

func checkIPLoop(config configuration) {
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
	List     []*config
	Interval time.Duration
}

func main() {
	log.SetPrefix("goDDNS: ")
	config := parseConfig()
	log.Println("launching goroutines")
	for _, c := range config.List {
		c.getIP = make(chan string)
		go c.listenIPLoop()
	}
	log.Println("launched goroutines")
	checkIPLoop(config)
}

//example config
//{
//"List": [
//{
//"Host": "@",
//"Domain": "aubble.com",
//"Password": "PASS",
//"Protocol": "namecheap"
//}
//],
//"Interval": 20 //seconds
//}
