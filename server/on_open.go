package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
)

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	var maxConns = bs.ConfigYml.GetInt("Server." + s.name + ".MaxConnections")
	var windowSize = bs.ConfigYml.GetInt("Server." + s.name + ".ReceiveWindowSize")
	var msg = fmt.Sprintf("[%s] OnOpen [%v<->%v]", s.name, c.RemoteAddr(), c.LocalAddr())
	var lc = s.LoginConns()
	if lc >= maxConns {
		// 已达到连接数阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlOfConns(lc, maxConns)...)
		return nil, gnet.Close
	} else if len(s.window) == windowSize {
		// 已达到接收窗口阈值时，拒绝新的连接
		log.Warn(msg, flowCtrlOfRecWindow(len(s.window), windowSize)...)
		return nil, gnet.Close
	} else {
		log.Info(msg, log.Int(ActiveConns, s.engine.CountConnections()), log.Int(CurrentWinSize, len(s.window)))
		return
	}
}

func flowCtrlOfConns(curVal, threshold int) (fields []log.Field) {
	fields = append(fields,
		FlowControl.Field(),
		TotalConnectionsThresholdReached.Field(),
		log.Int(CurrentValue, curVal),
		log.Int(Threshold, threshold),
	)
	return
}

func flowCtrlOfRecWindow(curVal, threshold int) (fields []log.Field) {
	fields = append(fields,
		FlowControl.Field(),
		TotalReceiveWindowsThresholdReached.Field(),
		log.Int(CurrentValue, curVal),
		log.Int(Threshold, threshold),
	)
	return
}

func (s *Server) LoginConns() int {
	counter := 0
	s.conns.Range(func(key, value interface{}) bool {
		counter++
		return true
	})
	return counter
}
