package xxc

import (
	"encoding/json"
	"log"
)

type Request struct {
	UserID int           `json:"userID"`
	Module string        `json:"module"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Data   interface{}   `json:"data"`
}

func (req *Request) MethodName() string {
	return req.Module + "." + req.Method
}

func (req *Request) String() string {
	v, err := json.Marshal(req)
	if err != nil {
		log.Printf("marshal request failed: %s:%s", req.Module, req.Method)
		return ""
	}
	return string(v)
}

type Response struct {
	Module  string          `json:"module"`
	Method  string          `json:"method"`
	Result  string          `json:"result"`
	Message string          `json:"message"`
	Params  []interface{}   `json:"params"`
	Data    json.RawMessage `json:"data"`
}

func (resp *Response) MethodName() string {
	return resp.Module + "." + resp.Method
}

func (resp *Response) String() string {
	v, err := json.Marshal(resp)
	if err != nil {
		log.Printf("marshal response failed: %s.%s", resp.Module, resp.Method)
		return ""
	}
	return string(v)
}

func parseResponse(data []byte) (*Response, error) {
	resp := &Response{}
	err := json.Unmarshal(data, resp)
	return resp, err
}

// XuanXuan Handler
type Handler interface {
	ServeXX(*Response)
}

type HandlerFunc func(*Response)

func (f HandlerFunc) ServeXX(r *Response) {
	f(r)
}

type UserProfile struct {
	Id       int    // ID
	Account  string // 用户名
	Realname string // 真实姓名
	Avatar   string // 头像URL
	Role     string // 角色
	Dept     int    // 部门ID
	Status   string // 当前状态
	Admin    string // 是否超级管理员，super 超级管理员 | no 普通用户
	Gender   string // 性别，u 未知 | f 女 | m 男
	Email    string // 邮箱
	Mobile   string // 手机
	Site     string // 网站
	Phone    string // 电话
	Signed   int64  // 登录时间
}

func parseUserProfile(data []byte) (*UserProfile, error) {
	profile := &UserProfile{}
	err := json.Unmarshal(data, profile)
	return profile, err
}
