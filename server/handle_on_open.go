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
  var msg = fmt.Sprintf("OnOpen [%v<->%v]", c.RemoteAddr(), c.LocalAddr())

  if s.LoginConns() >= maxConns {
    log.Warn(msg, log.String("option", "flow_control_of_total_connections"), log.Int("threshold", maxConns))
    return nil, gnet.Close
  } else if len(s.window) == windowSize {
    // 已达到窗口时，拒绝新的连接
    log.Warn(msg, log.String("option", "flow_control_of_total_receive_window"), log.Int("threshold", windowSize))
    return nil, gnet.Close
  } else {
    log.Info(msg, log.Int("active_connections", s.engine.CountConnections()), log.Int("current_window_size", len(s.window)))
    return
  }
}

func (s *Server) LoginConns() int {
  counter := 0
  s.conns.Range(func(key, value interface{}) bool {
    counter++
    return true
  })
  return counter
}
