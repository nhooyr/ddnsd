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

type Config struct {
	IP, Host, Domain, Password, Protocol, fqdn string
	getIP                                      chan string
}

func (c *Config) updateDNS() {
	dmsg := new(dns.Msg)
	dmsg.SetQuestion(c.fqdn, dns.TypeA)
	validIP := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	for c.IP = range c.getIP {
		in, err := dns.Exchange(dmsg, "dns1.registrar-servers.com:53")
		if err != nil {
			log.Println(err)
			continue
		}
		var curIP string
		if len(in.Answer) >= 1 {
			curIP = validIP.FindString(in.Answer[0].String())
			if curIP != c.IP {
				log.Println("updating", c.fqdn)
				var r *http.Response
				switch strings.ToLower(c.Protocol){
				case "namecheap":
					r, err = http.Get("https://dynamicdns.park-your-domain.com/update?host=" + c.Host + "&domain=" + c.Domain + "&password=" + c.Password + "&ip=" + c.IP)
				}
				if err != nil {
					log.Println(err)
					continue
				}
				defer r.Body.Close()
				buf, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println(err)
					continue
				}
				s := strings.ToLower(string(buf))
				i := strings.Index(s, "<responsestring>")
				if i != -1 {
					j := strings.Index(s, "</responsestring>")
					s = s[i+len("<responsestring>") : j]
				}
				if strings.Contains(s, "error") {
					if r.StatusCode != 200 {
						s = "problems connecting to server"
					}
					log.Printf("could not update %s because %s", c.fqdn, s)
					continue
				}
				log.Println("successfully updated", c.fqdn)
				curIP = c.IP
			} else {
				log.Println(c.fqdn, "already up to date")
			}
		} else {
			log.Println("cannot set ", c.fqdn)
		}
	}
}

func main() {
	log.SetPrefix("goDDNS: ")
	log.Println("reading config.json")
	f, err := os.Open("conf.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read config.json")
	log.Println("parsing config.json")
	configList := []*Config{}
	d := json.NewDecoder(f)
	err = d.Decode(&configList)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("parsed config.json")
	log.Println("launching goroutines")

	for _, c := range configList {
		if c.Host == "@" {
			c.fqdn = c.Domain + "."
		} else {
			c.fqdn = c.Host + "." + c.Domain + "."
		}
		c.getIP = make(chan string)
		go c.updateDNS()
	}
	log.Println("launched goroutines")

	for ;; time.Sleep(time.Second * 20) {
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
		IP := string(buf)
		log.Println("got", IP)
		log.Println("sending IP to goroutines")
		for _, c := range configList {
			c.getIP <- IP
		}
		log.Println("sent IP to goroutines")
	}
}
