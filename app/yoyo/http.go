package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/xjdrew/xxc"
)

var ErrUserOffline = errors.New("user is offline")

type SendMessageRequest struct {
	Users   []string `json:"users"`
	Group   string   `json:"group"`
	Message string   `json:"message"`
}

type GeneralResponse struct {
	Result  string `json:"result"` // "ok" or "failed"
	Message string `json:"code"`   // 错误描述
}

type myHttpHandler func(*http.Request) (interface{}, error)

func (f myHttpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("handle failed:", err)
		}
	}()

	statusCode := http.StatusOK
	var resp interface{}
	resp, err := f(req)
	if err != nil {
		resp = &GeneralResponse{
			Result:  "failed",
			Message: err.Error(),
		}
		statusCode = http.StatusBadRequest
	} else {
		if resp == nil {
			resp = &GeneralResponse{
				Result: "ok",
			}
		}
	}

	rw.Header().Set("Content-Type", "application/json") // normal header
	rw.WriteHeader(statusCode)

	chunk, err := json.Marshal(resp)
	if err != nil {
		log.Printf("ServeHTTP: %s", err)
	} else {
		rw.Write(chunk)
	}
}

type Server struct {
	Listen string
	user   atomic.Value
}

func parseTo(v interface{}, r *http.Request) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

func (s *Server) handleSendMessage(r *http.Request) (interface{}, error) {
	smr := &SendMessageRequest{}
	err := parseTo(smr, r)
	if err != nil {
		return nil, err
	}

	user := s.GetUser()
	if user == nil {
		return nil, ErrUserOffline
	}

	for _, account := range smr.Users {
		user.SayToUser(account, smr.Message)
	}

	if smr.Group != "" {
		user.SayToGroup(smr.Group, smr.Message)
	}
	return nil, nil
}

func (s *Server) handleGetUserList(r *http.Request) (interface{}, error) {
	user := s.GetUser()
	if user == nil {
		return nil, ErrUserOffline
	}
	return user.ReloadUserList(), nil
}

func (s *Server) SetUser(user *xxc.User) {
	s.user.Store(user)
}

func (s *Server) GetUser() *xxc.User {
	return s.user.Load().(*xxc.User)
}

func (s *Server) ListenAndServe() error {
	log.Println("listen:", s.Listen)
	http.Handle("/sendmessage", myHttpHandler(s.handleSendMessage))
	http.Handle("/getuserlist", myHttpHandler(s.handleGetUserList))
	return http.ListenAndServe(s.Listen, nil)
}
