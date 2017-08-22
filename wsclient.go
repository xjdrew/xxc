package xxc

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type call struct {
	resp *Response
	err  error
	done chan *call
}

type sessions struct {
	sync.Mutex
	m map[string]*call
}

func (ss *sessions) get(name string) *call {
	ss.Lock()
	defer ss.Unlock()
	return ss.m[name]
}

func (ss *sessions) set(name string, c *call) bool {
	ss.Lock()
	defer ss.Unlock()
	if _, ok := ss.m[name]; ok {
		return false
	}
	ss.m[name] = c
	return true
}

func (ss *sessions) delete(name string) {
	ss.Lock()
	defer ss.Unlock()
	delete(ss.m, name)
}

type wsClient struct {
	conn  *websocket.Conn
	token []byte

	rdMutex sync.Mutex
	wrMutex sync.Mutex

	ss *sessions
}

func (ws *wsClient) readMessage() (*Response, error) {
	ws.wrMutex.Lock()
	defer ws.wrMutex.Unlock()

	_, message, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	message, err = aesDecrypt(message, []byte(ws.token))
	if err != nil {
		return nil, err
	}

	return parseResponse(message)
}

func (ws *wsClient) writeMessage(req *Request) error {
	ws.rdMutex.Lock()
	defer ws.rdMutex.Unlock()

	s := req.String()
	data, err := aesEncrypt([]byte(s), []byte(ws.token))
	if err != nil {
		return err
	}
	err = ws.conn.WriteMessage(websocket.BinaryMessage, data)
	return err
}

func (ws *wsClient) handleMessage() {
	for {
		resp, err := ws.readMessage()
		if err != nil {
			log.Printf("read message failed: %s", err)
			return
		}
		name := resp.MethodName()
		call := ws.ss.get(name)
		if call == nil {
			log.Printf("unknown message: %s", name)
			continue
		}
		call.resp = resp
		call.done <- call
	}
}
func (ws *wsClient) Call(req *Request) (*Response, error) {
	name := req.MethodName()
	c := &call{
		done: make(chan *call),
	}

	if ok := ws.ss.set(name, c); !ok {
		return nil, fmt.Errorf("cannot call %s when previous call is pending", name)
	}
	defer ws.ss.delete(name)

	if err := ws.writeMessage(req); err != nil {
		return nil, err
	}

	<-c.done
	if c.err != nil {
		return nil, c.err
	}

	if c.resp.Result != "success" {
		return nil, fmt.Errorf("%s:%s", c.resp.Result, c.resp.Message)
	}
	return c.resp, nil
}

func createWsClient(httpUrl string, port int, token []byte) (*wsClient, error) {
	o, err := url.Parse(httpUrl)
	if err != nil {
		return nil, err
	}

	host := o.Host
	index := strings.LastIndex(host, ":")
	if index > 0 {
		host = host[:index]
	}
	host = fmt.Sprintf("%s:%d", host, port)
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	ws := &wsClient{
		conn:  conn,
		token: token,
		ss:    &sessions{m: make(map[string]*call)},
	}

	go ws.handleMessage()

	return ws, nil
}
