package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/xjdrew/xxc"
	"github.com/xjdrew/xxc/ext/tuling123"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

type TuringService struct {
	user   *xxc.User
	tuling *tuling123.Client
}

func (svc *TuringService) parseQuestion(message *xxc.ChatMessage) string {
	// 不处理非文本信息
	if message.ContentType != "text" {
		return ""
	}

	if message.User == svc.user.GetProfile().Id {
		return ""
	}

	group := svc.user.GetGroup(message.Cgid)
	tag := "@" + svc.user.GetProfile().Realname
	if group != nil && group.Type != "one2one" && strings.Index(message.Content, tag) == -1 {
		return ""
	}

	// remove tag
	content := message.Content
	for {
		index := strings.Index(content, tag)
		if index == -1 {
			break
		}
		content = content[:index] + content[index+len(tag):]
	}
	return content
}

func (svc *TuringService) handleOneMessage(message *xxc.ChatMessage) {
	question := svc.parseQuestion(message)
	if question == "" {
		return
	}

	log.Printf("\tquestion: %s", question)
	id := fmt.Sprintf("%d", svc.user.GetProfile().Id)
	tresp, err := svc.tuling.Ask(question, id)
	if err != nil {
		log.Printf("\ttuling answer failed: %s", err)
	}
	answer := tresp.String()
	log.Printf("\ttuling answer: %s", answer)
	svc.user.SayToGroup(message.Cgid, answer)
}

func (svc *TuringService) OnChatMessage(resp *xxc.Response) {
	var messages []*xxc.ChatMessage
	err := resp.ConvertDataTo(&messages)
	if err != nil {
		log.Printf("OnChatMessage failed: %s", err)
		return
	}

	log.Printf("OnChatMessage: <%s,%s>", resp.Result, resp.Message)
	for _, m := range messages {
		log.Printf("\tg:%s, u:%d, d:%d, t:%s, ct:%s, c:%s", m.Cgid, m.User, m.Date, m.Type, m.ContentType, m.Content)
		svc.handleOneMessage(m)
	}
}

func (svc *TuringService) Run() error {
	user := svc.user
	done := make(chan error)
	user.Client.Mux.HandleFunc("chat.message", svc.OnChatMessage)
	user.Client.Mux.HandleFunc("chat.kickoff", func(resp *xxc.Response) {
		done <- errors.New(resp.Message)
	})
	user.Client.HandleConnectionError(func(err error) {
		done <- err
	})
	return <-done
}

func main() {
	config := &xxc.ClientConfig{}

	var apiPath, apiKey string

	flag.StringVar(&config.Host, "host", "https://im.ejoy:11443", "http service")
	flag.StringVar(&config.User, "user", "bot", "user name")
	flag.StringVar(&config.Password, "password", "bot", "password")
	flag.BoolVar(&xxc.Verbose, "verbose", false, "print debug information")

	flag.StringVar(&apiPath, "apiPath", "http://www.tuling123.com/openapi/api", "tuling123 api interface")
	flag.StringVar(&apiKey, "apiKey", "83aedbb664414469a42adbccb03ae137", "tuling123 api key")
	flag.Parse()

	client := xxc.NewClient(config)
	user, err := xxc.CreateUser(client)
	if err != nil {
		log.Fatalf("create user failed: %s", err)
	}
	defer user.Fini()
	log.Printf("login as user: %s", user.GetProfile().Account)

	svc := &TuringService{
		user: user,
		tuling: &tuling123.Client{
			APIPath: apiPath,
			APIKey:  apiKey,
		},
	}

	log.Println(svc.Run())
}
