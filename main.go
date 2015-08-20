package main

import (
	"log"
	"os"
)

//TODO MORE PROTOCOLS

func main() {
	log.SetPrefix("goDDNS: ")
	os.Chdir("/etc/goDDNS")
	c := &configuration{logInterface: make(chan []interface{})}
	c.parseConfig()
	go c.receiveAndLog()
	c.log("launching goroutines")
	for _, d := range c.List {
		d.getIP = make(chan string)
		d.c = c
		go d.listenIPLoop()
	}
	c.log("launched goroutines")
	c.checkIPLoop()
}