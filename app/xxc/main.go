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

	var account, groupid, message string

	flag.StringVar(&config.Host, "host", "https://im.ejoy:11443", "http service")
	flag.StringVar(&config.User, "user", "bot", "user name")
	flag.StringVar(&config.Password, "password", "bot", "password")
	flag.BoolVar(&xxc.Verbose, "verbose", false, "print debug information")

	flag.StringVar(&account, "r", "", "指定接收信息的用户")
	flag.StringVar(&groupid, "g", "", "指定接收信息的组id")
	flag.StringVar(&message, "m", "", "消息的内容")
	flag.Parse()

	client := xxc.NewClient(config)
	user, err := xxc.CreateUser(client)
	if err != nil {
		log.Fatalf("create user failed: %s", err)
	}
	// defer user.Fini()

	log.Printf("login as user: %s", user.GetProfile().Account)

	if account != "" && message != "" {
		err := user.SayToUser(account, message)
		if err != nil {
			log.Printf("say failed: %s", err)
		}
	}

	if groupid != "" && message != "" {
		err := user.SayToGroup(groupid, message)
		if err != nil {
			log.Printf("say failed: %s", err)
		}
	}

	time.Sleep(100000 * time.Second)
}
