package main

import (
	"flag"
	"log"
	"time"

	"github.com/xjdrew/xxc"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	config := &xxc.ClientConfig{}

	flag.StringVar(&config.Host, "host", "https://im2.ejoy:11443", "http service")
	flag.StringVar(&config.User, "user", "bot", "user name")
	flag.StringVar(&config.Password, "password", "bot", "password")
	flag.BoolVar(&xxc.Verbose, "verbose", false, "print debug information")
	flag.Parse()

	c := xxc.NewClient(config)
	if err := c.Login(); err != nil {
		log.Fatalln(err)
	}
	log.Println("login succeed")
	err := c.Logout()
	log.Println(err)
	time.Sleep(100000 * time.Second)
}
