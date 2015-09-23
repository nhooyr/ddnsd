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
	LogPath      string
}

func (c *configuration) parseConfig(path string) {
	f, err := os.Open(path)
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

func (c *configuration) log(v ...interface{}) {
	logger.println(append([]interface{}{"--> global -/"}, v...))
}
