package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func (s *Server) OnClose(c gnet.Conn, e error) (action gnet.Action) {
	ses := Session(c)
	s.sessions.Delete(ses.Id())
	c.SetContext(nil)
	ses.conn = nil

	msg := fmt.Sprintf("[%s] OnClose [%v<->%v]", s.name, c.RemoteAddr(), c.LocalAddr())
	log.Warn(msg, ses.LogSid(), ConnectionClose.Field(), log.Int(LogKeyActiveConns, s.engine.CountConnections()), ErrorField(e))
	ses = nil
	return
}
