package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosms/my_errors"
)

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	var msg = fmt.Sprintf("[%s] OnOpen ===", s.name)

	if s.ActiveSessions() >= s.sessionPoolCap {
		// 已达到连接数阈值时，拒绝新的连接
		log.Warn(msg, FlatMapLog(s.LogCounter(), []log.Field{OpFlowControl.Field(), SErrField(my_errors.ErrorsSessionThreshReached)})...)
		return nil, gnet.Close
	} else {
		sc := s.CreateSession(c)
		log.Info(msg, FlatMapLog(sc.LogSession(), s.LogCounter())...)
		return
	}
}
