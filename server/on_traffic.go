package server

import (
	"github.com/panjf2000/gnet/v2"
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
type TrafficHandler func(cmd uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action)

// ExecuteChain all handlers
func ExecuteChain(handlers []TrafficHandler, cmd uint32, buff []byte, c gnet.Conn, s *Server) (action gnet.Action) {
	for _, handler := range handlers {
		handler := handler
		next, action := handler(cmd, buff, c, s)
		if next {
			continue
		} else {
			return action
		}
	}
	return action
}
