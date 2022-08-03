package sms

import (
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/hrygo/yaml_config"
	"golang.org/x/time/rate"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/client/session"
)

var smsConf yaml_config.YmlConfig
var factories [3]*SessionFactory

type SessionFactory struct {
	sync.Mutex
	srvName    string
	serverAddr string
	cli        *auth.Client
	sessions   []*session.Session
	window     chan struct{}
	limiter    *rate.Limiter
	regex      *regexp.Regexp
}

func init() {
	smsConf = yaml_config.CreateYamlFactory("config", "sms", bootstrap.ProjectName)
	smsConf.ConfigFileChangeListen()
}

// SelectSession 根据手机号码选择一个会话
func SelectSession(phone string) *session.Session {
	var fa *SessionFactory
	for i, factory := range factories {
		if factory == nil {
			factory = CreateSessionFactory(ISPS[i])
			factories[i] = factory
		}
		if factory.regex.MatchString(phone) {
			fa = factory
			break
		}
	}
	if fa == nil {
		return nil
	}
	return fa.PeekSession()
}

// CreateSessionFactory 创建或获取由isp指定的factory，isp需与sms.yml配置文件对应，否则会引起程序崩溃
func CreateSessionFactory(isp string) *SessionFactory {
	isp = strings.ToLower(isp)
	saved := factories[ISP(isp).Int()]
	if saved != nil {
		return saved
	}

	cli := auth.Cache.FindByCid(isp, smsConf.GetString(isp+".client-id"))
	if cli == nil {
		log.Fatalf("isp=%s, clientId=%s not found!", isp, smsConf.GetString(isp+".client-id"))
	}

	factory := &SessionFactory{srvName: isp, cli: cli}

	address := smsConf.GetString(isp + ".address")
	if address == "" {
		log.Fatal(isp + ".address can't be empty")
	}
	factory.serverAddr = address

	segment := smsConf.GetString(isp + ".segment")
	if segment == "" {
		log.Fatal(isp + ".segment can't be empty")
	}
	var err error
	factory.regex, err = regexp.Compile(segment)
	if err != nil {
		log.Fatal(err.Error())
	}

	maxConns := smsConf.GetInt(isp + ".max-conns")
	if maxConns > 0 {
		factory.sessions = make([]*session.Session, 0, maxConns)
	} else {
		factory.sessions = make([]*session.Session, 0, 2)
	}

	// 立即初始化一个连接
	c, err := net.Dial("tcp", address)
	if err == nil {
		sc := session.NewSession(isp, cli, c)
		if sc != nil {
			factory.sessions = append(factory.sessions, sc)
		}
	} else {
		log.Error(err.Error())
	}

	winSize := smsConf.GetInt(isp + ".mt-window-size")
	if maxConns > 0 {
		factory.window = make(chan struct{}, winSize)
	} else {
		factory.window = make(chan struct{}, 16)
	}

	// 默认1W微妙即10毫秒生成一个token，也即tps最大200
	ev := 10 * time.Millisecond
	throughput := smsConf.GetInt(isp + ".throughput")
	if throughput > 0 {
		// 1s = 1000*1000 microsecond = 1000000 microsecond, Throughput 单位时TPS
		ev = time.Duration(1000000/throughput) * time.Microsecond
	}
	limit := rate.Every(ev)
	factory.limiter = rate.NewLimiter(limit, cap(factory.window))
	factory.startTicker()

	factories[ISP(isp).Int()] = factory
	return factory
}

// PeekSession 获取排序后在头部的会话（最近最少使用的会话）
func (f *SessionFactory) PeekSession() *session.Session {
	if len(f.sessions) > 0 && f.sessions[0] == nil && f.sessions[0].HealthCheck() {
		return f.sessions[0]
	}

	var ret *session.Session
	for ret == nil || !ret.HealthCheck() {
		time.Sleep(time.Millisecond)
		if len(f.sessions) > 0 {
			ret = f.sessions[0]
		}
	}
	return ret
}

func (f *SessionFactory) startTicker() {
	go func() {
		d := smsConf.GetDuration(f.srvName + ".tick-duration")
		if d == 0 {
			d = time.Second
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			<-ticker.C
			f.onTick()
		}
	}()
}

func (f *SessionFactory) onTick() {
	maxConns := smsConf.GetInt(f.srvName + ".max-conns")
	var newSlice []*session.Session
	if maxConns <= 0 {
		maxConns = 2
	}
	newSlice = make([]*session.Session, 0, maxConns)

	f.Lock()
	sort.Sort(f) // 排序
	for _, sc := range f.sessions {
		if sc.HealthCheck() {
			sc.ResetCounter()               // 重置计数器
			newSlice = append(newSlice, sc) // 加入新会话列表
		} else { // 去除关闭无效会话
			sc.Close()
		}
	}
	// 使用新切片替换原列表
	f.sessions = newSlice
	f.Unlock()

	for len(f.sessions) < maxConns {
		f.newConnect()
		time.Sleep(time.Second)
	}
}

func (f *SessionFactory) Len() int {
	return len(f.sessions)
}

func (f *SessionFactory) Less(i, j int) bool {
	if i > len(f.sessions) {
		return true
	}
	if j > len(f.sessions) {
		return false
	}
	return f.sessions[i].LruPriority() > f.sessions[j].LruPriority()
}

func (f *SessionFactory) Swap(i, j int) {
	f.sessions[i], f.sessions[j] = f.sessions[j], f.sessions[i]
}

func (f *SessionFactory) newConnect() {
	// 立即初始化一个连接
	c, err := net.Dial("tcp", f.serverAddr)
	if err == nil {
		sc := session.NewSession(f.srvName, f.cli, c)
		if sc != nil {
			f.Lock()
			f.sessions = append(f.sessions, sc)
			f.Unlock()
		}
	} else {
		log.Error(err.Error())
	}
}

type ISP string

var ISPS = [3]string{"cmpp", "sgip", "smgp"}

func (i ISP) Int() int {
	switch i {
	case "cmpp":
		return 0
	case "sgip":
		return 2
	case "smgp":
		return 2
	}
	log.Panicf("ISP \"%s\" not found!", i)
	return -1
}
