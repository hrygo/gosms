package server

import (
	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Warnf("[%s] OnBoot @ %s", s.name, s.Address())
	s.engine = eng
	return
}
