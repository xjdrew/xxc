package main

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/xjdrew/xxc"
	"github.com/xjdrew/xxc/ext/tuling123"
)

type TuringService struct {
	user   atomic.Value
	Tuling tuling123.Client
}

func (svc *TuringService) parseQuestion(user *xxc.User, message *xxc.ChatMessage) string {
	// 不处理非文本信息
	if message.ContentType != "text" {
		return ""
	}

	if message.User == user.GetProfile().Id {
		return ""
	}

	group := user.GetGroup(message.Cgid)
	tag := "@" + user.GetProfile().Realname
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
	user := svc.GetUser()
	if user == nil {
		log.Println("no user")
		return
	}

	question := svc.parseQuestion(user, message)
	if question == "" {
		return
	}

	log.Printf("\tquestion: %s", question)
	id := fmt.Sprintf("%d", message.User)
	tresp, err := svc.Tuling.Ask(question, id)
	if err != nil {
		log.Printf("\ttuling answer failed: %s", err)
		return
	}
	answer := tresp.String()
	log.Printf("\ttuling answer: %s", answer)
	user.SayToGroup(message.Cgid, answer)
}

func (svc *TuringService) OnChatMessage(resp *xxc.Response) {
	var messages []*xxc.ChatMessage
	err := resp.ConvertDataTo(&messages)
	if err != nil {
		log.Printf("OnChatMessage failed: %s", err)
		return
	}

	if resp.Succeed() {
		for _, m := range messages {
			log.Printf("OnChatMessage: g<%s>, u<%d>, d<%d>, t<%s>, ct<%s>, c<%s>", m.Cgid, m.User, m.Date, m.Type, m.ContentType, m.Content)
			svc.handleOneMessage(m)
		}
	} else {
		log.Printf("OnChatMessage: <%s,%s>", resp.Result, resp.Message)
	}
}

func (svc *TuringService) SetUser(user *xxc.User) {
	if svc.Tuling.APIKey == "" || svc.Tuling.APIPath == "" {
		return
	}
	svc.user.Store(user)

	if user != nil {
		user.Client.Mux.HandleFunc("chat.message", svc.OnChatMessage)
	}

}

func (svc *TuringService) GetUser() *xxc.User {
	return svc.user.Load().(*xxc.User)
}
