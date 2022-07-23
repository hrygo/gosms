package server

import (
	"fmt"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
)

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnTick @ %s://:%d", s.name, s.protocol, s.port)
	log.Info(msg)

	s.conns.Range(func(key, value interface{}) bool {
		addr := key.(string)
		con, ok := value.(gnet.Conn)
		if ok {
			_ = s.pool.Submit(func() {
				msg = fmt.Sprintf("[%s] OnTick [%v<->%v] active test.", s.name, addr, con.LocalAddr())
				// var active Packet
				switch s.name {
				case CMPP:
					// active =
				case SGIP:
					// active =
				case SMGP:
					// active =
				}
				// err := con.AsyncWrite(active.Encode(), nil)
				// if err == nil {
				//    log.Info(msg, ActiveTest.Field(), log.Reflect("packet", active))
				// } else {
				//   log.Error(msg, ActiveTest.Field(),  ErrorField(err))
				// }
			})
		}
		return true
	})
	return bs.ConfigYml.GetDuration("Server.TickDuration"), gnet.None
}
