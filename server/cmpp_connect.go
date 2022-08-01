package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/codec/cmpp"
)

var cmppConnect TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_CONNECT) != cmd {
		return true, gnet.None
	}

	login := &cmpp.Connect{}
	sc := Session(c)
	err := login.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(sc, buff)
		return false, gnet.Close
	}

	// 异步处理登录逻辑，避免阻塞 event-loop
	err = s.GoPool().Submit(func() {
		handleCmppConnect(s, sc, login)
	})
	if err != nil {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC),
			FlatMapLog(sc.LogSession(), []log.Field{OpConnectionClose.Field(), ErrorField(err), Packet2HexLogStr(buff)})...)
		return false, gnet.Close
	}

	return false, gnet.None
}

// 注意：登录异常时，发送响应后，可直接关闭连接，此时无法传递 gnet.Action 了
func handleCmppConnect(s *Server, sc *session, login *cmpp.Connect) {
	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)
	// 打印登录报文
	log.Info(msg, FlatMapLog(sc.LogSession(16), login.Log())...)

	// 获取客户端信息
	cli := auth.Cache.FindByCid(s.name, login.SourceAddr())
	code := login.Check(cli)

	// 检查当前已登录会话数是否已达上限
	if code == cmpp.ConnStatusOK {
		// 注意这里仅按照单节点计算某个client的session数，实际上应该计算集群中的某个client的session数。
		// TODO 要支持集群，连接会话的计数，应该采用数据库或者redis等存储。
		activeSession := s.CountSessionByClientId(cli.ClientId)
		if activeSession >= cli.MaxConns {
			code = cmpp.ConnStatusOthers
		}
	}

	resp := login.ToResponse(uint32(code))
	pack := resp.Encode()
	err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		if code == cmpp.ConnStatusOK {
			sc.completeLogin(cli)
			// 更新会话
			s.SessionPool().Store(sc.id, sc)
		} else {
			// 客户端登录失败，关闭连接
			sc.stat = StatClosing
			_ = c.Flush()
			_ = c.Close()
		}
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Info(msg, FlatMapLog(sc.LogSession(16), resp.Log())...)
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_CONNECT.OpLog(), SErrField(err.Error())})...)
	}
}
