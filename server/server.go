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

// 定义运营商常量
const (
	CMPP = "CMPP"
	SMGP = "SMGP"
	SGIP = "SGIP"
)

const (
	protocol  = "tcp"
	LogErrKey = "error"

	CurrentValue   = "curVal"
	Threshold      = "threshold"
	ActiveConns    = "active_connections"
	CurrentWinSize = "current_receive_windows_size"
)

// Server 封装 gnet server
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
		defer s.pool.Release()
		addr := fmt.Sprintf("%s://:%d", s.protocol, s.port)
		err := gnet.Run(
			s,
			addr,
			gnet.WithTicker(true),
			gnet.WithMulticore(s.multicore),
		)
		log.Fatalf("server(%s) exits with error: %v", addr, err)
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
	return &Server{
		protocol:  protocol,
		port:      port,
		multicore: multicore,
		pool:      pool,
		window:    make(chan struct{}, receiveWindowSize), // 用通道控制消息接收窗口
		name:      name,
	}
}

func (s *Server) Engine() gnet.Engine {
	return s.engine
}

func (s *Server) Protocol() string {
	return s.protocol
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Pool() *goroutine.Pool {
	return s.pool
}

func (s *Server) Conns() *sync.Map {
	return &s.conns
}

func (s *Server) Window() chan struct{} {
	return s.window
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) Address() string {
	return fmt.Sprintf("%s://:%d", s.protocol, s.port)
}

// Operation 定义操作类型
type Operation byte

const (
	operation = "op" // 操作类型

	FlowControl Operation = iota // 操作类型枚举
	ConnectionClose
	ActiveTest
)

func (op Operation) String() string {
	return []string{
		"flow_control",
		"connection_close",
		"active_test",
	}[op-1]
}

func (op Operation) Field() log.Field {
	return log.String(operation, op.String())
}

// Reason 操作对应的具体原因
type Reason byte

const (
	reason = "reason" // 操作原因

	TotalConnectionsThresholdReached Reason = iota // 操作原因类型枚举
	TotalReceiveWindowsThresholdReached
)

func (op Reason) String() string {
	return []string{
		"total_connections_threshold_reached",
		"total_receive_window_threshold_reached",
	}[op-1]
}

func (op Reason) Field() log.Field {
	return log.String(operation, op.String())
}

func ErrorField(err error) log.Field {
	return log.Reflect(LogErrKey, err)
}
