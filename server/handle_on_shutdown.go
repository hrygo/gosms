package server

import (
  "fmt"
  "time"

  "github.com/hrygo/log"
  "github.com/panjf2000/gnet/v2"
)

func (s *Server) OnShutdown(eng gnet.Engine) {
  addr := fmt.Sprintf("%s://:%d", s.protocol, s.port)
  log.Warnf("OnShutdown [%s] @ %s ...", s.name, addr)
  for eng.CountConnections() > 0 {
    log.Warnf("OnShutdown [%s] @ %s active connections is %d, waiting...", s.name, addr, eng.CountConnections())
    time.Sleep(10 * time.Millisecond)
  }
  log.Warnf("OnShutdown [%s] @ %s Completed.", s.name, addr)
}
