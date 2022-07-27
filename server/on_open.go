package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/my_errors"
)

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	var maxConns = bs.ConfigYml.GetInt("Server." + s.name + ".MaxConnections")
	var windowSize = bs.ConfigYml.GetInt("Server." + s.name + ".ReceiveWindowSize")
	var msg = fmt.Sprintf("[%s] OnOpen ===", s.name)

	var lc = s.engine.CountConnections()
	if lc >= maxConns {
		// 已达到连接数阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlLogFields(s, nil, SErrField(my_errors.ErrorsConnsThreshReached))...)
		return nil, gnet.Close
	} else if len(s.window) == windowSize {
		// 已达到接收窗口阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlLogFields(s, nil, SErrField(my_errors.ErrorsTotalRWinThreshReached))...)
		return nil, gnet.Close
	} else {
		ses := s.SaveSession(c)
		log.Info(msg, FlatMapLog(ses.LogSession(), s.LogCounter())...)
		return
	}
}

func flowCtrlLogFields(s *Server, ses *session, reason ...log.Field) (fields []log.Field) {
	return FlatMapLog(ses.LogSession(), s.LogCounter(), []log.Field{OpFlowControl.Field()}, reason)
}
