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

	var groupid, message string

	flag.StringVar(&config.Host, "host", "https://im2.ejoy:11443", "http service")
	flag.StringVar(&config.User, "user", "bot", "user name")
	flag.StringVar(&config.Password, "password", "bot", "password")
	flag.BoolVar(&xxc.Verbose, "verbose", false, "print debug information")

	flag.StringVar(&groupid, "groupid", "", "指定接收信息的组id")
	flag.StringVar(&message, "message", "", "消息的内容")
	flag.Parse()

	client := xxc.NewClient(config)
	user, err := CreateUser(client)
	if err != nil {
		log.Fatalf("create user failed: %s", err)
	}
	defer user.Fini()

	log.Printf("login as user: %s", user.GetProfile().Account)

	if groupid != "" && message != "" {
		err := user.SayToGroup(groupid, message)
		if err != nil {
			log.Printf("say failed: %s", err)
		}
	}

	time.Sleep(100000 * time.Second)
}
