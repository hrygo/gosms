package server

import (
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func (s *Server) OnShutdown(eng gnet.Engine) {
	log.Warnf("[%s] OnShutdown @ %s ...", s.name, s.Address())
	for eng.CountConnections() > 0 {
		log.Warnf("[%s] OnShutdown @ %s active connections is %d, waiting...", s.name, s.Address(), eng.CountConnections())
		time.Sleep(10 * time.Millisecond)
	}
	log.Warnf(" [%s] OnShutdown @ %s Completed.", s.name, s.Address())
}
