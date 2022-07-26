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

	msg := fmt.Sprintf("[%s] OnClose ===", s.name)
	fields := JoinLog(SSR(ses, c.RemoteAddr()), ConnectionClose.Field(), ErrorField(e))
	log.Warn(msg, JoinLog(fields, CCWW(s)...)...)
	ses = nil
	return
}
