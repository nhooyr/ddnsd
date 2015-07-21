package main

import "log"

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