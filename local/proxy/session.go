package proxy

import (
	"log"
	"sync"

	"github.com/yinqiwen/gsnova/common/event"
)

// import "net"

var sessions map[uint32]*ProxySession = make(map[uint32]*ProxySession)
var sessionMutex sync.Mutex
var sessionNotExist error

type ProxySession struct {
	id          uint32
	queue       *event.EventQueue
	Channel     ProxyChannel
	Hijacked    bool
	SSLHijacked bool
}

func (s *ProxySession) handle(ev event.Event) error {
	s.queue.Publish(ev)
	return nil
}

func (s *ProxySession) Close() error {
	closeEv := &event.TCPCloseEvent{}
	closeEv.SetId(s.id)
	s.handle(closeEv)
	return nil
}

func getProxySession(sid uint32) *ProxySession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	s, exist := sessions[sid]
	if exist {
		return s
	}
	return nil
}

func newProxySession(sid uint32, queue *event.EventQueue) *ProxySession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	s := new(ProxySession)
	s.id = sid
	s.queue = queue
	sessions[s.id] = s
	log.Printf("Create proxy session:%d", sid)
	return s
}

func closeProxySession(sid uint32) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessions, sid)
	log.Printf("Close proxy session:%d, %d left", sid, len(sessions))
}

func HandleEvent(ev event.Event) error {
	session := getProxySession(ev.GetId())
	if nil == session {
		log.Printf("No session:%d found for %T", ev.GetId(), ev)
		return sessionNotExist
	}
	return session.handle(ev)
}
