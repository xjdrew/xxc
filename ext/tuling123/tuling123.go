package tuling123

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Request struct {
	Key    string `json:"key"`
	Info   string `json:"info"`
	Loc    string `json:"loc"`
	Userid string `json:"userid"`
}

type ResponseListItem struct {
	Article   string
	Source    string
	Icon      string
	Detailurl string
}

type Response struct {
	Code int
	Text string
	Url  string
	List []ResponseListItem
}

func (r *Response) String() string {
	switch r.Code {
	case 200000:
		return fmt.Sprintf("%s\n%s", r.Text, r.Url)
	case 302000:
		var items []string
		for _, item := range r.List {
			items = append(items, fmt.Sprintf("* [%s【%s】](%s)", item.Article, item.Source, item.Detailurl))
		}
		return fmt.Sprintf("%s\n%s", r.Text, strings.Join(items, "\n"))
	default:
		return r.Text
	}
}

// http://www.tuling123.com/help/h_cent_webapi.jhtml?nav=doc
type Client struct {
	APIPath string
	APIKey  string
}

func (c *Client) Ask(question string, id string) (*Response, error) {
	treq := &Request{
		Key:    c.APIKey,
		Info:   question,
		Userid: id,
	}
	data, err := json.Marshal(treq)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	resp, err := http.Post(c.APIPath, "application/json", reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	tresp := &Response{}
	if err := dec.Decode(tresp); err != nil {
		return nil, err
	}
	return tresp, nil
}
