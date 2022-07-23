package server

import (
  "fmt"
  "sync"
  "time"

  "github.com/hrygo/log"
  "github.com/panjf2000/ants/v2"
  "github.com/panjf2000/gnet/v2"
  "github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

  bs "github.com/hrygo/gosmsn/bootstrap"
)

const (
  CMPP     = "CMPP"
  SMGP     = "SMGP"
  SGIP     = "SGIP"
  protocol = "tcp"
)

type Server struct {
  gnet.BuiltinEventEngine
  engine    gnet.Engine
  protocol  string
  port      int
  multicore bool
  pool      *goroutine.Pool
  conns     sync.Map
  window    chan struct{}
  name      string
}

func Start(s *Server) {
  go func() {
    addr := fmt.Sprintf("%s://:%d", s.protocol, s.port)
    err := gnet.Run(
      s,
      addr,
      gnet.WithTicker(true),
      gnet.WithMulticore(s.multicore),
    )
    log.Errorf("server(%s) exits with error: %v", addr, err)
  }()
}

func New(name string) *Server {
  var port = bs.ConfigYml.GetInt("Server." + name + ".Port")
  var multicore = bs.ConfigYml.GetBool("Server." + name + ".Multicore")
  var maxPoolSize = bs.ConfigYml.GetInt("Server." + name + ".MaxPoolSize")
  var receiveWindowSize = bs.ConfigYml.GetInt("Server." + name + ".ReceiveWindowSize")
  var options = ants.Options{
    ExpiryDuration:   time.Minute, // 1 分钟内不被使用的worker会被清除
    Nonblocking:      false,       // 如果为true,worker池满了后提交任务会直接返回nil
    MaxBlockingTasks: maxPoolSize, // blocking模式有效，否则worker池满了后提交任务会直接返回nil
    PreAlloc:         false,
    PanicHandler: func(e interface{}) {
      log.Errorf("%v", e)
    },
  }
  var pool, _ = ants.NewPool(maxPoolSize, ants.WithOptions(options))
  defer pool.Release()
  return &Server{
    protocol:  protocol,
    port:      port,
    multicore: multicore,
    pool:      pool,
    window:    make(chan struct{}, receiveWindowSize), // 用通道控制消息接收窗口
    name:      name,
  }
}
