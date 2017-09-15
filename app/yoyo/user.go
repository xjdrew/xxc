package main

import (
	"errors"
	"log"
	"time"

	"github.com/xjdrew/xxc"
)

type XXUser struct {
	Config xxc.ClientConfig

	Online  func(user *xxc.User)
	Offline func(user *xxc.User)
}

// 喧喧服务器有bug：
//		如果登录时踢人下线，其他客户端可能先收到用户登录，后马上又收到用户下线通知，导致状态不对
// 		快速的登录并退出一次，确保状态干净
func (xxu *XXUser) flash() {
	client := xxc.NewClient(&xxu.Config)
	user, err := xxc.CreateUser(client)
	if err != nil {
		return
	}
	user.Fini()
}

func (xxu *XXUser) Serve() error {
	done := make(chan error)
	for {
		client := xxc.NewClient(&xxu.Config)
		user, err := xxc.CreateUser(client)
		if err != nil {
			log.Fatalf("create user failed: %s", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("%s online", xxu.Config.User)
		client.Mux.HandleFunc("chat.kickoff", func(resp *xxc.Response) {
			select {
			case done <- errors.New(resp.Message):
			default:
			}
		})

		client.HandleConnectionError(func(err error) {
			select {
			case done <- err:
			default:
			}
		})

		if xxu.Online != nil {
			xxu.Online(user)
		}

		log.Printf("%s offline: %s", xxu.Config.User, <-done)

		if xxu.Offline != nil {
			xxu.Offline(user)
		}

		user.Fini()

		// 等待10秒再登录
		time.Sleep(5 * time.Second)
		xxu.flash()
		time.Sleep(1 * time.Second)
	}
}
