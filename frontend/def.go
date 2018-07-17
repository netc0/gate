package frontend

import (
	"log"
	"sync"
	"github.com/netc0/gate/protocol"
	"sync/atomic"
)

// define
var (
	g_sessionId int32 = 0
)

func (this* TCPSession) IsOk() bool {
	return this.isOk
}

var (
	sessions = make(map[string]protocol.ISession)
	sessionMutex = new(sync.Mutex)

	onNewSession func(sid string)
	onCloseSession func(sid string)
)

func GetSession(sid string) protocol.ISession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	var s = sessions[sid]
	return s
}

// 清空会话
func ClearSession(owner interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	for k, v := range sessions {
		if v.GetOwner() == owner {
			delete(sessions, k)
		}
	}
}

// 遍历会话
func ForeachSession(callback func(session protocol.ISession)) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	for _, v := range sessions {
		callback(v)
	}
}
// 添加会话Frontend启动
func AddSession(inst interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	session, ok := inst.(protocol.ISession)
	if ok {
		//log.Println("新连接进入", session.GetId(), session.GetOwner())
		sessions[session.GetId()] = session
		return
	}

	log.Println("cast to ISession error")
}
// 删除会话
func RemoveSession(inst interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	session, ok := inst.(protocol.ISession)
	if ok {
		//log.Println("Frontend 关闭会话", session.GetId())
		delete(sessions, session.GetId())
	}
}

func NewSessionId() int32 {
	atomic.AddInt32(&g_sessionId, 1)
	return g_sessionId
}