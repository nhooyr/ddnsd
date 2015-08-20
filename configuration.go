package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type configuration struct {
	List         []*domain
	Interval     time.Duration
	Logfile      string
	logInterface chan []interface{}
}

func (c *configuration) parseConfig() {
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	d := json.NewDecoder(f)
	err = d.Decode(c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *configuration) checkIPLoop() {
	for ; ; time.Sleep(time.Second * c.Interval) {
		newIP, err := c.getPublicIP()
		if err != nil {
			c.log(err)
			continue
		}
		c.log("got", newIP)
		c.log("sending IP to goroutines")
		for _, d := range c.List {
			d.getIP <- newIP
		}
		c.log("sent IP to goroutines")
		c.log("sleeping checkIPLoop for", int64(c.Interval))
	}
}

func (c *configuration) getPublicIP() (string, error) {
	c.log("getting public IP")
	resp, err := http.Get("https://api.ipify.org")
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

// receive on global log channel and append received to Logfile
// if Logfile doesn't exist, create it, and check continuously
// if it doesn't exist and if so create
func (c *configuration) receiveAndLog() {
	if c.Logfile != "" {
		for {
			logFile, err := os.OpenFile(c.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer logFile.Close()
			log.SetOutput(logFile)
			log.Println("--> global -/ beginning logging")
			for {
				log.Println(<-c.logInterface...)
				if _, err = os.Stat(c.Logfile); os.IsNotExist(err) {
					break
				}
			}
		}
	}
}

func (c *configuration) log(v ...interface{}) {
	c.logInterface <- append([]interface{}{"--> global -/"}, v...)
}