package xxc

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type httpClient struct {
	*http.Client
	host string
}

func urlJoin(s string, v string) string {
	if strings.HasSuffix(s, "/") {
		return s + v
	}
	return s + "/" + v
}

func (c *httpClient) DoHttpJsonRequest(path string, data string, ret interface{}) error {
	v := url.Values{}
	v.Set("data", data)
	resp, err := c.Client.PostForm(urlJoin(c.host, path), v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(ret)
}

func newHttpClient(host string) *httpClient {
	var client *http.Client
	if host[:6] != "https:" {
		client = &http.Client{}
	} else {
		client = &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	}
	return &httpClient{
		Client: client,
		host:   host,
	}
}
