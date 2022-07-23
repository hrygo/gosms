package server

import (
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec/cmpp"
)

var cmppSubmit TrafficHandler = func(cmd uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_SUBMIT) != cmd {
		return true, gnet.None
	}

	return
}
