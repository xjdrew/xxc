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
	loginErr   error
	user       *UserProfile

	Mux *ClientMux
}

func (c *Client) getMux() *ClientMux {
	if c.Mux == nil {
		return DefaultClientMux
	}
	return c.Mux
}

func (c *Client) initWsClient() error {
	ws, err := createWsClient(c.clientConfig.Host, c.serverConfig.ChatPort, []byte(c.serverConfig.Token), c.getMux())
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

	resp, err := c.Call(loginReq)
	if err != nil {
		return err
	}

	profile, err := parseUserProfile(resp.Data)
	if err != nil {
		return err
	}

	c.user = profile

	if Verbose {
		log.Printf("user profile: %+v", c.user)
	}
	return nil
}

func (c *Client) Login() error {
	if err := c.init(); err != nil {
		return err
	}

	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()

	if c.user != nil {
		return nil
	}

	if c.loginErr != nil {
		return c.loginErr
	}

	if err := c.loginWithLocked(); err != nil {
		c.loginErr = err
		return err
	}
	return nil
}

func (c *Client) Logout() error {
	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()
	if c.user == nil {
		return nil
	}

	logoutReq := &Request{
		UserID: c.user.Id,
		Module: "chat",
		Method: "logout",
	}

	_, err := c.Call(logoutReq)
	if err != nil {
		log.Printf("logout failed: %s", err)
		return err
	}
	c.user = nil
	return nil
}

func (c *Client) GetUser() (*UserProfile, error) {
	err := c.Login()
	if err != nil {
		return nil, err
	}
	return c.user, nil
}

func (c *Client) Call(req *Request) (*Response, error) {
	if Verbose {
		log.Printf("+ call %s", req.MethodName())
	}
	resp, err := c.wsClient.Call(req)

	if Verbose {
		if err != nil {
			log.Printf("- call return %s: %s", req.MethodName(), err)
		} else {
			log.Printf("- call return %s: %s", req.MethodName(), resp)
		}
	}
	return resp, err
}

func (c *Client) Send(req *Request) error {
	if Verbose {
		log.Printf("* send %s", req.MethodName())
	}
	err := c.wsClient.Send(req)
	return err
}

func NewClient(config *ClientConfig) *Client {
	if Verbose {
		log.Printf("clientConfig: %+v", config)
	}
	return &Client{
		clientConfig: config,
	}
}
