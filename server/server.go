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
	protocol = "tcp"

	// 定义运营商常量

	CMPP = "CMPP"
	SMGP = "SMGP"
	SGIP = "SGIP"
)

// Server 封装 gnet server
type Server struct {
	sync.Mutex
	gnet.BuiltinEventEngine
	engine    gnet.Engine
	protocol  string
	port      int
	multicore bool
	pool      *goroutine.Pool
	sessions  sync.Map
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

func (s *Server) SaveSession(c gnet.Conn) *session {
	ses := createSession(c)
	c.SetContext(ses)
	s.sessions.Store(ses.id, ses)
	return ses
}

func (s *Server) CountConnsByClientId(clientId string) (counter uint32) {
	s.Lock()
	defer s.Unlock()
	s.sessions.Range(func(key, value any) bool {
		s, ok := value.(*session)
		if ok && s.clientId == clientId {
			counter += 1
		}
		return true
	})
	return
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
	return &s.sessions
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

func Session(c gnet.Conn) *session {
	ctx := c.Context()
	if ctx == nil {
		return nil
	}
	ses, ok := ctx.(*session)
	if ok {
		return ses
	}
	return nil
}
