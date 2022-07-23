package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func sgipOnTraffic(s *Server, c gnet.Conn) (action gnet.Action) {
	var msg = fmt.Sprintf("[%s] OnTraffic [%v<->%v]", s.name, c.RemoteAddr(), c.LocalAddr())
	log.Debug(msg)
	return
}
