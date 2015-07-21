package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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