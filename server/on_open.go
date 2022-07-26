package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
)

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	var maxConns = bs.ConfigYml.GetInt("Server." + s.name + ".MaxConnections")
	var windowSize = bs.ConfigYml.GetInt("Server." + s.name + ".ReceiveWindowSize")
	var msg = fmt.Sprintf("[%s] OnOpen ===", s.name)

	var lc = s.engine.CountConnections()
	if lc >= maxConns {
		// 已达到连接数阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlLogFields(s, nil, c, TotalConnectionsThresholdReached.Field())...)
		return nil, gnet.Close
	} else if len(s.window) == windowSize {
		// 已达到接收窗口阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlLogFields(s, nil, c, TotalReceiveWindowsThresholdReached.Field())...)
		return nil, gnet.Close
	} else {
		ses := s.SaveSession(c)
		log.Info(msg, JoinLog(SSR(ses, c.RemoteAddr()), CCWW(s)...)...)
		return
	}
}

func flowCtrlLogFields(s *Server, ses *session, c gnet.Conn, reason log.Field) (fields []log.Field) {
	fields = SSR(ses, c.RemoteAddr())
	fields = append(fields, CCWW(s)...)
	fields = append(fields, FlowControl.Field(), reason)
	return
}
