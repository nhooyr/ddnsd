package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

//TODO MORE PROTOCOLS

// global filelogger
var logger *fileLogger

func main() {
	log.SetPrefix("ddnsd: ")
	// flag variables
	var (
		stderr, errPrefix bool
		config            string
	)
	flag.BoolVar(&stderr, "e", false, "stderr logging")
	flag.BoolVar(&errPrefix, "t", false, "stderr logging prefix (name, timestamp)")
	flag.StringVar(&config, "c", "/usr/local/etc/ddnsd/config.json", "path to configuration file")
	flag.Parse()
	dir, _ := filepath.Split(config)
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			log.Fatalln("--> global -/", err)
		}
	}
	c := new(configuration)
	c.parseConfig(config)
	logger = &fileLogger{stderr: stderr}
	if errPrefix == false {
		log.SetFlags(0)
		log.SetPrefix("")
	}
	if c.LogPath != "" {
		logFile, err := os.OpenFile(c.LogPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		logger.Logger = log.New(logFile, "ddnsd: ", 3)
	}
	c.log("launching goroutines")
	for _, d := range c.List {
		d.getIP = make(chan string)
		d.c = c
		go d.listenIPLoop()
	}
	c.log("launched goroutines")
	c.checkIPLoop()
}
