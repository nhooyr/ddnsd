package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

//TODO MORE PROTOCOLS

func main() {
	log.SetPrefix("goDDNS: ")
	var path string
	flag.StringVar(&path, "c", "/etc/goDDNS/config.json", "path to configuration file")
	flag.Parse()
	dir, _ := filepath.Split(path)
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			log.Fatalln("--> global -/", err)
		}
	}
	c := &configuration{logInterface: make(chan []interface{})}
	c.parseConfig(path)
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
