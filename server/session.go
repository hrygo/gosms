package server

import (
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec"
)

// 会话信息 gnet.Conn 的附加属性
type session struct {
	sync.Mutex
	id          uint64
	conn        gnet.Conn
	clientId    string    // 客户端识别号，由服务端分配
	ver         byte      // 协议版本号
	stat        stat      // 会话状态
	nAt         byte      // 未接收到响应的心跳次数
	lastUseTime time.Time // 接收到客户端的 active/active_resp 或 mt 消息会更新该时间
	counter               // mt, dly, report 计数器
	createTime  time.Time
}

type counter struct {
	mt, dly, report uint64 //  接收到的下行短信、发送的上行短信、发送的状态报告的数量
}

// 会话状态
type stat byte

const (
	StatConnect stat = iota
	StatLogin
	StatClosing
)

func createSession(conn gnet.Conn) *session {
	se := &session{}
	se.id = uint64(codec.B64Seq.NextVal())
	se.conn = conn
	se.createTime = time.Now()
	se.lastUseTime = time.Now()
	return se
}

func (s *session) Conn() gnet.Conn {
	return s.conn
}

func (s *session) Id() uint64 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *session) ClientId() string {
	return s.clientId
}

func (s *session) Ver() byte {
	return s.ver
}

func (s *session) Stat() stat {
	return s.stat
}

func (s *session) NAt() byte {
	return s.nAt
}

func (s *session) CreateTime() time.Time {
	return s.createTime
}

func (s *session) LastUseTime() time.Time {
	return s.lastUseTime
}

func (s *session) LogSid() log.Field {
	return log.Uint64(LogKeySessionId, s.Id())
}
