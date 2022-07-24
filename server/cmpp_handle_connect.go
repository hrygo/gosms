package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec/cmpp"
)

var cmppConnect TrafficHandler = func(cmd uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_CONNECT) != cmd {
		return true, gnet.None
	}

	login := &cmpp.ConnReqPkt{}
	err := login.Unpack(buff)
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
func handleCmppConnect(s *Server, c gnet.Conn, login *cmpp.ConnReqPkt) {
	var msg = fmt.Sprintf("[%s] OnTraffic_HandleConnect [%v<->%v]", s.Name(), c.RemoteAddr(), c.LocalAddr())
	log.Info(msg, log.Reflect(LogKeyPacket, login))

	// TODO 登录检查
	Session(c).stat = StatLogin

	// TODO  发送响应
}
