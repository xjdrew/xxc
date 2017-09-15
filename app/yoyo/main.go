package main

import (
	"flag"
	"log"

	"github.com/xjdrew/xxc"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	xxu := &XXUser{}
	turing := &TuringService{}
	server := &Server{}

	flag.StringVar(&xxu.Config.Host, "host", "https://im.ejoy:11443", "http service")
	flag.StringVar(&xxu.Config.User, "user", "bot", "user name")
	flag.StringVar(&xxu.Config.Password, "password", "bot", "password")
	flag.BoolVar(&xxc.Verbose, "verbose", false, "print debug information")

	flag.StringVar(&turing.Tuling.APIPath, "apiPath", "http://www.tuling123.com/openapi/api", "tuling123 api interface")
	flag.StringVar(&turing.Tuling.APIKey, "apiKey", "", "tuling123 api key")

	flag.StringVar(&server.Listen, "listen", ":1954", "http listen address")
	flag.Parse()

	xxu.Online = func(user *xxc.User) {
		turing.SetUser(user)
		server.SetUser(user)
	}

	xxu.Offline = func(user *xxc.User) {
		turing.SetUser(nil)
		server.SetUser(nil)
	}

	done := make(chan error)
	go func() {
		done <- server.ListenAndServe()
	}()

	go func() {
		done <- xxu.Serve()
	}()

	log.Printf("finish: %s", <-done)
}
