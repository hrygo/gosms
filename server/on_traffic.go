package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/my_errors"
)

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	switch s.name {
	case CMPP:
		return cmppOnTraffic(s, c)
	case SGIP:
		return sgipOnTraffic(s, c)
	case SMGP:
		return smgpOnTraffic(s, c)
	}
	return gnet.Close
}

// TrafficHandler Don't Read From gnet.Conn.
// cmd: command id ;
// seq: request sequence ;
// buff: packet body ;
// next: true continue next handler, false return the action
type TrafficHandler func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action)

// ExecuteChain all handlers
func ExecuteChain(handlers []TrafficHandler, cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (action gnet.Action) {
	for _, handler := range handlers {
		handler := handler
		next, action := handler(cmd, seq, buff, c, s)
		if next {
			continue
		} else {
			return action
		}
	}
	return action
}

func sessionCheck(s *session) bool {
	if s.stat != StatLogin {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", s.ServerName(), RC),
			FlatMapLog(s.LogSession(), []log.Field{OpConnectionClose.Field(), SErrField(my_errors.ErrorsDecodePacketBody)})...)
		return false
	}
	return true
}

func decodeErrorLog(s *session, buff []byte) {
	log.Error(fmt.Sprintf("[%s] OnTraffic %s", s.ServerName(), RC),
		FlatMapLog(s.LogSession(), []log.Field{OpConnectionClose.Field(), SErrField(my_errors.ErrorsDecodePacketBody), Packet2HexLogStr(buff)})...)
}
