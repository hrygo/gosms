package server

import (
  "fmt"

  "github.com/hrygo/log"
  "github.com/panjf2000/gnet/v2"
)

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
  addr := fmt.Sprintf("%s://:%d", s.protocol, s.port)
  log.Warnf("OnBoot [%s] @ %s ...", s.name, addr)
  s.engine = eng
  return
}
