package server

import (
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"golang.org/x/time/rate"

	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec"
)

// 会话信息 gnet.Conn 的附加属性
type session struct {
	sync.Mutex
	id          uint64
	createTime  time.Time
	conn        gnet.Conn
	clientId    string          // 客户端识别号，由服务端分配
	serverName  string          // 连接的Server的name
	ver         byte            // 协议版本号
	stat        stat            // 会话状态
	nAt         byte            // 未接收到响应的心跳次数
	lastUseTime time.Time       // 接收到客户端的 active/active_resp 或 mt 消息会更新该时间
	counter                     // mt, dly, report 计数器
	window      chan struct{}   // 流控所需通道，登录成功后需设置此值，否则消息不能正常收发
	pool        *goroutine.Pool // 会话级别的线程池，登录成功后需设置此值，否则消息不能正常收发
	limiter     *rate.Limiter   // 限速器
}

type counter struct {
	mt, dly, report uint64 //  接收到的下行短信、发送的上行短信、发送的状态报告的数量
}

// 会话状态
type stat byte

const (
	StatConnect stat = iota
	StatLogin
	StatClosing
)

func NewSession(c gnet.Conn) *session {
	se := &session{}
	se.id = uint64(codec.B64Seq.NextVal())
	se.stat = StatConnect
	se.conn = c
	se.createTime = time.Now()
	se.lastUseTime = time.Now()
	return se
}

func createSessionSidePool(size int) *goroutine.Pool {
	var options = ants.Options{
		ExpiryDuration:   time.Minute, // 1 分钟内不被使用的worker会被清除
		Nonblocking:      false,       // 如果为true,worker池满了后提交任务会直接返回nil
		MaxBlockingTasks: size,        // blocking模式有效，否则worker池满了后提交任务会直接返回nil
		PreAlloc:         false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	var pool, _ = ants.NewPool(size, ants.WithOptions(options))
	return pool
}

func (s *session) completeLogin(cli *client.Client) {
	// 设置会话信息及会话级别资源，此代码非常重要！！！
	s.Lock()
	defer s.Unlock()
	s.stat = StatLogin
	s.ver = cli.Version
	s.clientId = cli.ClientId
	s.lastUseTime = time.Now()
	s.closeResource()
	s.window = make(chan struct{}, cli.MtWindowSize)
	s.pool = createSessionSidePool(cli.MtWindowSize * 2)
	s.setupLimiter(cli)
}

func (s *session) setupLimiter(cli *client.Client) {
	// 默认1W微妙即10毫秒生成一个token，也即tps最大200
	ev := 10 * time.Millisecond
	if cli != nil && cli.Throughput != 0 {
		// 1s = 1000*1000 microsecond = 1000000 microsecond, Throughput 单位时TPS
		ev = time.Duration(1000000/cli.Throughput) * time.Microsecond
	}
	limit := rate.Every(ev)
	s.limiter = rate.NewLimiter(limit, cap(s.window))
}

// 关闭通道和线程池
func (s *session) closeResource() {
	if s == nil {
		return
	}

	if s.window != nil {
		close(s.window)
		s.window = nil
	}
	if s.pool != nil {
		s.pool.Release()
		s.pool = nil
	}
	s.limiter = nil
}

func (s *session) Window() chan struct{} {
	return s.window
}

func (s *session) Pool() *goroutine.Pool {
	return s.pool
}

func (s *session) Conn() gnet.Conn {
	return s.conn
}

func (s *session) Id() uint64 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *session) ClientId() string {
	return s.clientId
}

func (s *session) Ver() byte {
	return s.ver
}

func (s *session) Stat() stat {
	return s.stat
}

func (s *session) NAt() byte {
	return s.nAt
}

func (s *session) CreateTime() time.Time {
	return s.createTime
}

func (s *session) LastUseTime() time.Time {
	return s.lastUseTime
}

func (s *session) Limiter() *rate.Limiter {
	return s.limiter
}

func (s *session) LogSession(size ...int) []log.Field {
	if s == nil {
		return nil
	}
	var ret []log.Field
	if len(size) == 0 {
		ret = make([]log.Field, 0, 8)
	} else if size[0] > 8 {
		ret = make([]log.Field, 0, size[0])
	}
	cliName := s.clientId
	if cliName == "" {
		cliName = "not login"
	}
	remote := "closed"
	if s.conn != nil {
		remote = s.conn.RemoteAddr().String()
	}
	return append(ret,
		log.String(SrvName, s.serverName),
		log.String(CliName, cliName),
		log.Uint64(Sid, s.Id()),
		log.String(RemoteAddr, remote))
}

func (s *session) LogCounter() []log.Field {
	var mt, dlv, rpt uint64 = 0, 0, 0
	if s != nil {
		mt, dlv, rpt = s.counter.mt, s.counter.dly, s.counter.report
	}
	cli := client.Cache.FindByCid(s.serverName, s.clientId)
	if cli == nil {
		cli = &client.Client{}
	}

	// 防止未登录而未初始化pool时的空指针异常。
	free := 0
	pcap := 0
	if s.pool != nil {
		free = s.pool.Free()
		pcap = s.pool.Cap()
	}

	return []log.Field{
		log.Int(LogKeyClientConnsCap, cli.MaxConns),
		log.Int(LogKeySessionPoolFree, free),
		log.Int(LogKeySessionPoolCap, pcap),
		log.Int(LogKeySessionSwCur, len(s.window)),
		log.Int(LogKeySessionSwCap, cap(s.window)),
		log.Uint64(LogKeyCounterMt, mt),
		log.Uint64(LogKeyCounterDlv, dlv),
		log.Uint64(LogKeyCounterRpt, rpt),
	}
}

func (s *session) ServerName() string {
	return s.serverName
}

func (s *session) CounterAddMt() {
	s.Lock()
	defer s.Unlock()
	s.mt += 1
	s.lastUseTime = time.Now()
}

func (s *session) CounterAddDly() {
	s.Lock()
	defer s.Unlock()
	s.dly += 1
	// 当前模拟上行短信由自身触发，不算客户端活动，不更新时间
	// s.lastUseTime = time.Now()
}

func (s *session) CounterAddRpt() {
	s.Lock()
	defer s.Unlock()
	s.report += 1
	s.lastUseTime = time.Now()
}
