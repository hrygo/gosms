package server

import (
	"fmt"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec/cmpp"
)

var cmppConnect TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_CONNECT) != cmd {
		return true, gnet.None
	}

	login := &cmpp.Connect{}
	err := login.Decode(seq, buff)
	if err != nil {
		return false, gnet.Close
	}

	// 异步处理登录逻辑，避免阻塞 event-loop
	err = s.Pool().Submit(func() {
		handleCmppConnect(s, c, login)
	})

	return false, gnet.None
}

// 注意：登录异常时，发送响应后，可直接关闭连接，此时无法传递 gnet.Action 了
func handleCmppConnect(s *Server, c gnet.Conn, login *cmpp.Connect) {
	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)

	session := Session(c)
	log.Info(msg, JoinLog(SSR(session, c.RemoteAddr(), 16), login.Log()...)...)

	// 获取客户端信息
	cli := client.Cache.FindByCid(s.name, login.SourceAddr())
	code := login.Check(cli)
	resp := login.ToResponse(uint32(code))

	// send cmpp_connect_resp async
	_ = s.pool.Submit(func() {
		pack := resp.Encode()
		err := c.AsyncWrite(pack, func(c gnet.Conn) error {
			msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
			log.Info(msg, JoinLog(SSR(session, c.RemoteAddr(), 16), resp.Log()...)...)

			if code == cmpp.ConnStatusOK {
				session.Lock()
				defer session.Unlock()
				session.ver = byte(login.Version())
				session.stat = StatLogin
				session.lastUseTime = time.Now()
				s.Conns().Store(session.id, session)
			} else {
				// 客户端登录失败，关闭连接
				session.stat = StatClosing
				_ = c.Close()
			}
			return nil
		})
		if err != nil {
			log.Error(msg, JoinLog(SSR(session, c.RemoteAddr()), ErrorField(err))...)
		}
	})
}
