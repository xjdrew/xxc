package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/satori/go.uuid"
	"github.com/xjdrew/xxc"
)

type User struct {
	client      *xxc.Client
	profile     *xxc.UserProfile
	loginFinish chan struct{} // 登录完成后关闭

	usersMutex sync.RWMutex
	users      map[int]*xxc.UserProfile // 所有用户

	groupMutex sync.RWMutex
	groups     map[string]*xxc.ChatGroup // 所有会话
}

func (u *User) updateUsers(users []*xxc.UserProfile) {
	u.usersMutex.Lock()
	defer u.usersMutex.Unlock()

	if u.users == nil {
		u.users = make(map[int]*xxc.UserProfile)
	}
	for _, user := range users {
		u.users[user.Id] = user
	}
}

func (u *User) updateGroups(groups []*xxc.ChatGroup) {
	u.groupMutex.Lock()
	defer u.groupMutex.Unlock()

	if u.groups == nil {
		u.groups = make(map[string]*xxc.ChatGroup)
	}

	for _, group := range groups {
		u.groups[group.Gid] = group
		println("---------------- gid:", group.Gid, group.Name)
	}
}

func (u *User) GetGroup(gid string) *xxc.ChatGroup {
	u.groupMutex.RLock()
	defer u.groupMutex.RUnlock()
	return u.groups[gid]
}

// 刷新用户列表
func (u *User) OnChatUserGetList(resp *xxc.Response) {
	var users []*xxc.UserProfile
	err := resp.ConvertDataTo(&users)
	if err != nil {
		log.Printf("OnChatUserGetList failed: %s", err)
		return
	}
	log.Printf("OnChatUserGetList: %d users", len(users))
	u.updateUsers(users)
}

// 刷新回话列表
func (u *User) OnChatGetList(resp *xxc.Response) {
	var groups []*xxc.ChatGroup
	err := resp.ConvertDataTo(&groups)
	if err != nil {
		log.Printf("OnChatGetList failed: %s", err)
		return
	}
	log.Printf("OnChatGetList: %d groups", len(groups))
	u.updateGroups(groups)

	// 登录完成
	close(u.loginFinish)
}

//
// 刷新回话列表
func (u *User) OnChatMessage(resp *xxc.Response) {
	var messages []*xxc.ChatMessage
	err := resp.ConvertDataTo(&messages)
	if err != nil {
		log.Printf("OnChatMessage failed: %s", err)
		return
	}

	log.Printf("OnChatMessage: %d messages", len(messages))
	for _, m := range messages {
		log.Printf("\tg:%s, u:%d, d:%d, t:%s, ct:%s, c:%s", m.Cgid, m.User, m.Date, m.Type, m.ContentType, m.Content)
	}
}

func (u *User) say(gid string, content string) error {
	message := &xxc.ChatMessage{
		Gid:         uuid.NewV4().String(),
		Cgid:        gid,
		Type:        "normal",
		ContentType: "text",
		Date:        0,
		User:        u.profile.Id,
		Content:     content,
	}

	var params struct {
		Messages []*xxc.ChatMessage `json:"messages"`
	}
	params.Messages = []*xxc.ChatMessage{
		message,
	}

	messageRequest := &xxc.Request{
		UserID: u.profile.Id,
		Module: "chat",
		Method: "message",
		Params: params,
	}
	return u.client.Send(messageRequest)
}

func (u *User) SayToGroup(gid string, content string) error {
	if u.GetGroup(gid) == nil {
		return fmt.Errorf("%s is not a group id", gid)
	}
	return u.say(gid, content)
}

func (u *User) GetProfile() *xxc.UserProfile {
	return u.profile
}

func (u *User) Fini() error {
	return u.client.Logout()
}

func CreateUser(client *xxc.Client) (*User, error) {
	user := &User{
		loginFinish: make(chan struct{}),
	}
	mux := &xxc.ClientMux{}
	mux.HandleFunc("chat.usergetlist", user.OnChatUserGetList)
	mux.HandleFunc("chat.getlist", user.OnChatGetList)
	mux.HandleFunc("chat.message", user.OnChatMessage)
	client.Mux = mux

	profile, err := client.GetUser()
	if err != nil {
		return nil, err
	}

	user.client = client
	user.profile = profile

	// 登录成功后，会依次收到三条消息: chat.usergetlist, chat.getlist, chat.message
	<-user.loginFinish
	return user, nil
}
