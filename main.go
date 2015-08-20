package main

import (
	"log"
	"os"
)

//TODO MORE PROTOCOLS

func main() {
	log.SetPrefix("goDDNS: ")
	os.Chdir("/etc/goDDNS")
	c := configuration{logInterface: make(chan []interface{})}
	c.parseConfig()
	c.receiveAndLog()
	c.log("launching goroutines")
	for _, c := range c.List {
		c.getIP = make(chan string)
		go c.listenIPLoop()
	}
	c.log("launched goroutines")
	c.checkIPLoop()
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
