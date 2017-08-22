package xxc

import (
	"encoding/json"
	"log"
)

type Request struct {
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
	Module  string        `json:"module"`
	Method  string        `json:"method"`
	Result  string        `json:"result"`
	Message string        `json:"message"`
	Params  []interface{} `json:"params"`
	Data    interface{}   `json:"data"`
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
