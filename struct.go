package xxc

import (
	"encoding/json"
	"log"
)

type Request struct {
	UserID int         `json:"userID"`
	Module string      `json:"module"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Data   interface{} `json:"data"`
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
	Params  interface{}     `json:"params"`
	Data    json.RawMessage `json:"data"`
}

func (resp *Response) MethodName() string {
	return resp.Module + "." + resp.Method
}

func (resp *Response) ConvertDataTo(v interface{}) error {
	return json.Unmarshal(resp.Data, v)
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
	Signed   int64  // ???
}

type ChatGroup struct {
	Id             int         // 会话在服务器数据保存的id
	Gid            string      // 会话的全局id,
	Name           string      // 会话的名称
	Type           string      // 会话的类型
	Admins         string      // 会话允许发言的用户列表
	Committers     string      //
	Subject        int         // 主题会话的关联主题ID
	Public         int         // 是否公共会话
	CreatedBy      string      // 创建者用户名
	CreatedDate    int64       // 创建时间
	EditedBy       string      // 编辑者用户名
	EditedDate     interface{} // 编辑时间 ----------------  这个字段服务器有bug，有值时是整数，但默认值居然是空字符串
	LastActiveTime int64       // 会话最后一次发送消息的时间
	Star           int         // 当前登录用户是否收藏此会话
	Hide           int         // 当前登录用户是否隐藏此会话
	Mute           int         //
	Members        []int       // 当前会话中包含的所有用户信息,只需要包含id即可
}

type ChatMessage struct {
	Id          int    `json:"-"`           // 消息在服务器保存的id
	Gid         string `json:"gid"`         // 此消息的gid
	Cgid        string `json:"cgid"`        // 此消息关联的会话的gid
	User        int    `json:"user"`        // 消息发送的用户ID
	Date        int64  `json:"date"`        // 消息发送的时间
	Type        string `json:"type"`        // 消息的类型
	ContentType string `json:"contentType"` // 消息内容的类型
	Content     string `json:"content"`     // 消息内容
}
