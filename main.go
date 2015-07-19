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
)

//TODO MORE PROTOCOLS

type config struct {
	IP, Host, Domain, Password, Protocol, fqdn string
	getIP                                      chan string
}

func (c *config) listenIPLoop() {
	if c.Host == "@" {
		c.fqdn = c.Domain + "."
	} else {
		c.fqdn = c.Host + "." + c.Domain + "."
	}
	for newIP := range c.getIP {
		err := c.updateIP(newIP)
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *config) useProtocol() (r *http.Response, err error) {
	switch strings.ToLower(c.Protocol) {
	case "namecheap":
		r, err = http.Get("https://dynamicdns.park-your-domain.com/update?host=" + c.Host + "&domain=" + c.Domain + "&password=" + c.Password + "&ip=" + c.IP)
	}
	return
}

func (c *config) checkError(buf []byte, r *http.Response) (errStr string, bad bool) {
	s := (string(buf))
	if strings.Contains(s, "error") {
		bad = true
		i := strings.Index(s, "<ResponseString>")
		j := strings.Index(s, "</ResponseString>")
		s = s[i+len("<ResponseString>") : j]
		errStr = fmt.Sprintf("could not update %s %s\nstatus code %s", c.fqdn, s, r.StatusCode)
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
	log.Println("successfully updated", c.fqdn)
	return nil
}

func parseConfig()(configList []*config){
	log.Println("reading config.json")
	f, err := os.Open("/etc/goDDNS/config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read config.json")
	log.Println("parsing config.json")
	configList = []*config{}
	d := json.NewDecoder(f)
	err = d.Decode(&configList)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("parsed config.json")
	return
}

func checkIPLoop(configList []*config){
	oldIP := ""
	for ;; time.Sleep(time.Second * 5) {
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
		if oldIP != newIP {
			log.Println("sending IP to goroutines")
			for _, c := range configList {
				c.getIP <- newIP
			}
			log.Println("sent IP to goroutines")
			oldIP = newIP
		} else {
			log.Print("up to date")
		}
	}
}

func main() {
	log.SetPrefix("goDDNS: ")
	configList := parseConfig()
	log.Println("launching goroutines")
	for _, c := range configList {
		c.getIP = make(chan string)
		go c.listenIPLoop()
	}
	log.Println("launched goroutines")
	checkIPLoop(configList)
}
