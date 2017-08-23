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
		println("------:", user.Account, user.Realname)
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
		log.Printf("group: gid:%s name:%s type:%s", group.Gid, group.Name, group.Type)
	}
}

func (u *User) GetGroup(gid string) *xxc.ChatGroup {
	u.groupMutex.RLock()
	defer u.groupMutex.RUnlock()
	return u.groups[gid]
}

// 查找用户所在的 one2one 会话组
func (u *User) QueryOne2OneGroup(id int) *xxc.ChatGroup {
	u.groupMutex.RLock()
	defer u.groupMutex.RUnlock()

	for _, group := range u.groups {
		if group.Type == "one2one" {
			for _, member := range group.Members {
				if member == id {
					return group
				}
			}
		}
	}
	return nil
}

// 创建一个 one2one Group
func (u *User) CreateOne2OneGroup(id int) (gid string, err error) {
	gid = fmt.Sprintf("%d&%d", u.profile.Id, id)
	createGroupRequest := &xxc.Request{
		UserID: u.profile.Id,
		Module: "chat",
		Method: "create",
		Params: []interface{}{
			gid,
			"",
			"one2one",
			[]int{u.profile.Id, id},
			0,
			false,
		},
	}

	err = u.client.Send(createGroupRequest)
	return
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

// 刷新会话列表
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
// 接收聊天信息
func (u *User) OnChatMessage(resp *xxc.Response) {
	var messages []*xxc.ChatMessage
	err := resp.ConvertDataTo(&messages)
	if err != nil {
		log.Printf("OnChatMessage failed: %s", err)
		return
	}

	log.Print("OnChatMessage")
	for _, m := range messages {
		log.Printf("\tg:%s, u:%d, d:%d, t:%s, ct:%s, c:%s", m.Cgid, m.User, m.Date, m.Type, m.ContentType, m.Content)
	}
}

// 接收新建组信息
func (u *User) OnChatCreate(resp *xxc.Response) {
	var group *xxc.ChatGroup
	err := resp.ConvertDataTo(&group)
	if err != nil {
		log.Printf("OnChatCreate failed: %s", err)
		return
	}

	log.Printf("OnChatCreate: groupid:%s, name:%s, type:%s", group.Gid, group.Name, group.Type)
	u.updateGroups([]*xxc.ChatGroup{group})
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
		return fmt.Errorf("%s is not a valid group id", gid)
	}
	return u.say(gid, content)
}

// 通过用户account，查找用户
func (u *User) GetUserByAccount(account string) *xxc.UserProfile {
	u.usersMutex.RLock()
	defer u.usersMutex.RUnlock()
	for _, user := range u.users {
		if user.Account == account {
			return user
		}
	}
	return nil
}

func (u *User) SayToUser(account string, content string) error {
	user := u.GetUserByAccount(account)
	if user == nil {
		return fmt.Errorf("%s is not a valid account", account)
	}
	group := u.QueryOne2OneGroup(user.Id)
	if group != nil {
		return u.say(group.Gid, content)
	}

	// 走到这里，说明这两个人之前没有私聊过
	// 先创建一个 one2one Group
	gid, err := u.CreateOne2OneGroup(user.Id)
	if err != nil {
		return err
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
	mux.HandleFunc("chat.create", user.OnChatCreate)
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
