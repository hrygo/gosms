package server

import (
	"fmt"
	"runtime"
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

	CMPP = "cmpp"
	SMGP = "smgp"
	SGIP = "sgip"
)

// Server 封装 gnet server
type Server struct {
	sync.Mutex
	gnet.BuiltinEventEngine
	engine         gnet.Engine
	name           string
	protocol       string
	port           int
	multicore      bool
	goPool         *goroutine.Pool // 处理登录请求的Pool
	sessionPool    sync.Map        // 存储会话的map（GoPool）
	sessionPoolCap int             // 存储会话的map的最大容量
	activeSessions int             // 已存储的会话数，活跃会话数
}

func Start(s *Server) {
	go func() {
		defer s.goPool.Release()
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
	var MaxSessions = bs.ConfigYml.GetInt("Server." + name + ".MaxSessions")
	var options = ants.Options{
		ExpiryDuration: time.Minute, // 1 分钟内不被使用的worker会被清除
		Nonblocking:    true,        // 如果为true, worker池满了后提交任务会直接返回nil，如果为false需设置MaxBlockingTasks参数为非0值
		PreAlloc:       false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	// 因为该pool目前仅用于处理登录请求，不需过大，设置与CPU核心数相同。
	var pool, _ = ants.NewPool(runtime.NumCPU(), ants.WithOptions(options))
	return &Server{
		name:           name,
		protocol:       protocol,
		port:           port,
		multicore:      multicore,
		goPool:         pool,
		sessionPoolCap: MaxSessions,
	}
}

func (s *Server) CreateSession(c gnet.Conn) *session {
	s.Lock()
	defer s.Unlock()
	sc := NewSession(c)
	sc.serverName = s.name
	c.SetContext(sc)
	s.sessionPool.Store(sc.id, sc)
	s.activeSessions += 1
	return sc
}

func (s *Server) CountSessionByClientId(clientId string) (counter int) {
	s.Lock()
	defer s.Unlock()
	s.sessionPool.Range(func(key, value any) bool {
		s, ok := value.(*session)
		if ok && s.clientId == clientId {
			counter += 1
		}
		return true
	})
	return
}

func (s *Server) ActiveSessions() int {
	return s.activeSessions
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

func (s *Server) GoPool() *goroutine.Pool {
	return s.goPool
}

func (s *Server) SessionPool() *sync.Map {
	return &s.sessionPool
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) SessionPoolCap() int {
	return s.sessionPoolCap
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

func (s *Server) LogCounter() []log.Field {
	return []log.Field{
		log.Int(LogKeyPoolFree, s.SessionPoolCap()-s.ActiveSessions()),
		log.Int(LogKeyPoolCap, s.SessionPoolCap()),
	}
}
func (s *Server) LogCounterWithName() []log.Field {
	fields := s.LogCounter()
	ret := make([]log.Field, 0, len(fields)+1)
	ret = append(ret, log.String(SrvName, s.name))
	return append(ret, fields...)
}
