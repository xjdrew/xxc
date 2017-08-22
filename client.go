package xxc

import (
	"log"
	"sync"
)

var Verbose = false

type ClientConfig struct {
	Host     string
	User     string
	Password string
}

type ServerConfig struct {
	Version        string `json:"version"`
	Token          string `json:"token"`
	SiteType       string `json:"siteType"`
	UploadFileSize int64  `json:"uploadFileSize"`
	ChatPort       int    `json:"chatPort"`
	TestModel      bool   `json:"testModel"`
}

type Client struct {
	clientConfig *ClientConfig
	serverConfig *ServerConfig

	httpClient *httpClient
	wsClient   *wsClient

	initMutex sync.Mutex
	initErr   error

	loginMutex sync.Mutex
	logined    bool
	loginErr   error
}

func (c *Client) initWsClient() error {
	ws, err := createWsClient(c.clientConfig.Host, c.serverConfig.ChatPort, []byte(c.serverConfig.Token))
	if err != nil {
		return err
	}
	c.wsClient = ws
	return nil
}

func (c *Client) initWithLocked() error {
	if Verbose {
		log.Printf("init start")
	}

	serverConfigReq := &Request{
		Module: "chat",
		Method: "login",
		Params: []interface{}{
			"",
			c.clientConfig.User,
			hashPassword(c.clientConfig.Password),
			"online",
		},
	}

	serverConfig := &ServerConfig{}

	c.httpClient = newHttpClient(c.clientConfig.Host)
	err := c.httpClient.DoHttpJsonRequest("serverInfo", serverConfigReq.String(), serverConfig)
	if err != nil {
		c.initErr = err
		return err
	}

	c.serverConfig = serverConfig

	if Verbose {
		log.Printf("serverConfig:%+v", serverConfig)
		log.Printf("init end")
	}
	return nil
}

func (c *Client) init() error {
	c.initMutex.Lock()
	defer c.initMutex.Unlock()

	if c.initErr != nil {
		return c.initErr
	}

	if c.serverConfig != nil {
		return nil
	}
	return c.initWithLocked()
}

func (c *Client) loginWithLocked() error {
	if Verbose {
		log.Printf("login start")
	}
	if err := c.initWsClient(); err != nil {
		c.loginErr = err
		return err
	}

	loginReq := &Request{
		Module: "chat",
		Method: "login",
		Params: []interface{}{
			"",
			c.clientConfig.User,
			hashPassword(c.clientConfig.Password),
			"online",
		},
	}

	resp, err := c.wsClient.Call(loginReq)
	if err != nil {
		return err
	}

	if Verbose {
		log.Printf("login response: %s", resp)
		log.Printf("login start")
	}
	return nil
}

func (c *Client) Login() error {
	if err := c.init(); err != nil {
		return err
	}

	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()

	if c.logined {
		return nil
	}

	if c.loginErr != nil {
		return c.loginErr
	}

	if err := c.loginWithLocked(); err != nil {
		c.loginErr = err
		return err
	}
	c.logined = true
	return nil
}

func (c *Client) Logout() error {
	return nil
}

func NewClient(config *ClientConfig) *Client {
	if Verbose {
		log.Printf("clientConfig: %+v", config)
	}
	return &Client{
		clientConfig: config,
	}
}
