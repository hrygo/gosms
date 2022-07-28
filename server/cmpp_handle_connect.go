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
	ses := Session(c)
	err := login.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(ses, buff)
		return false, gnet.Close
	}

	// 异步处理登录逻辑，避免阻塞 event-loop
	_ = s.Pool().Submit(func() {
		handleCmppConnect(s, ses, login)
	})

	return false, gnet.None
}

// 注意：登录异常时，发送响应后，可直接关闭连接，此时无法传递 gnet.Action 了
func handleCmppConnect(s *Server, sc *session, login *cmpp.Connect) {
	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)

	// 打印登录报文
	log.Info(msg, FlatMapLog(sc.LogSession(16), login.Log())...)

	// 获取客户端信息
	cli := client.Cache.FindByCid(s.name, login.SourceAddr())
	// cli 为空检查
	code := login.Check(cli)
	resp := login.ToResponse(uint32(code))
	pack := resp.Encode()
	err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Info(msg, FlatMapLog(sc.LogSession(16), resp.Log())...)

		if code == cmpp.ConnStatusOK {
			// 设置会话信息及会话级别资源，此代码非常重要！！！
			sc.Lock()
			defer sc.Unlock()
			sc.ver = byte(login.Version())
			sc.stat = StatLogin
			sc.clientId = cli.ClientId
			sc.lastUseTime = time.Now()
			sc.closePoolChan()
			sc.window = make(chan struct{}, cli.MtWindowSize)
			sc.pool = createSessionSidePool(cli.MtWindowSize * 2)
			s.Conns().Store(sc.id, sc)
		} else {
			// 客户端登录失败，关闭连接
			sc.stat = StatClosing
			_ = c.Close()
		}
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_CONNECT.Log(), SErrField(err.Error())})...)
	}
}
