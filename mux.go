package xxc

import (
	"log"
	"sync"
)

type ClientMux struct {
	mu sync.RWMutex
	m  map[string]Handler
}

func (mux *ClientMux) HandleFunc(methodName string, handler func(*Response)) {
	mux.Handle(methodName, HandlerFunc(handler))
}

func (mux *ClientMux) Handle(methodName string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if methodName == "" {
		panic("mux: invalid methodName " + methodName)
	}

	if handler == nil {
		panic("mux: nil handler")
	}

	if mux.m == nil {
		mux.m = make(map[string]Handler)
	}

	if _, ok := mux.m[methodName]; ok {
		panic("mux: multiple registrations for " + methodName)
	}
	mux.m[methodName] = handler
}

func (mux *ClientMux) ServeXX(resp *Response) {
	if Verbose {
		log.Printf("message %s: %s", resp.MethodName(), resp)
	}

	mux.mu.RLock()
	defer mux.mu.RUnlock()
	h, ok := mux.m[resp.MethodName()]
	if !ok {
		log.Printf("message %s: no handler", resp.MethodName())
	} else {
		h.ServeXX(resp)
	}
}

var DefaultClientMux = &ClientMux{}
