package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func (s *Server) OOnClose(c gnet.Conn, e error) (action gnet.Action) {
	s.conns.Delete(c.RemoteAddr().String())

	msg := fmt.Sprintf("[%s] OnClose [%v<->%v]", s.name, c.RemoteAddr(), c.LocalAddr())
	log.Warn(msg, ConnectionClose.Field(), log.Int(CurrentValue, s.LoginConns()), ErrorField(e))
	return
}
