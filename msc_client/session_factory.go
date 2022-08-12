package sms

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"golang.org/x/time/rate"

	"github.com/hrygo/gosms/event_manager"
	"github.com/hrygo/gosms/smc_client/session"

	"github.com/hrygo/gosms/auth"
	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

var factories [3]*SessionFactory

// resultQueryCacheMap 临时存储短信发送的返回结果数据，Key为queryId,value为[]*Status，后续采用数据库存储
var resultQueryCacheMap sync.Map

var mu sync.Mutex

func AsyncPool() *goroutine.Pool {
	return pool
}

type SessionFactory struct {
	sync.Mutex
	srvName    string
	serverAddr string
	authConf   *codec.AuthConf
	sessions   []*session.Session
	window     chan struct{}
	limiter    *rate.Limiter
	regex      *regexp.Regexp
}

// SelectSession 根据手机号码选择一个会话
func SelectSession(phone string) *session.Session {
	mu.Lock()
	var fa *SessionFactory
	for i, factory := range factories {
		if factory == nil {
			factory = CreateSessionFactory(ISPS[i])
			if factory == nil {
				continue
			}
		}
		if factory.regex.MatchString(phone) {
			fa = factory
			break
		}
	}
	mu.Unlock()

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

	ac := FindAuthConf(isp, ConfigYml.GetString(isp+".client-id"))
	if ac == nil {
		log.Errorf("isp=%s, clientId=%s not found!", isp, ConfigYml.GetString(isp+".client-id"))
		return nil
	}

	factory := &SessionFactory{srvName: isp, authConf: ac}

	address := ConfigYml.GetString(isp + ".address")
	if address == "" {
		log.Error(isp + ".address can't be empty")
	}
	factory.serverAddr = address

	segment := ConfigYml.GetString(isp + ".segment")
	if segment == "" {
		log.Error(isp + ".segment can't be empty")
	}
	var err error
	factory.regex, err = regexp.Compile(segment)
	if err != nil {
		log.Error(err.Error())
	}

	maxConns := ConfigYml.GetInt(isp + ".max-conns")
	if maxConns > 0 {
		factory.sessions = make([]*session.Session, 0, maxConns)
	} else {
		factory.sessions = make([]*session.Session, 0, 2)
	}

	// 立即初始化一个连接
	c, err := net.Dial("tcp", address)
	if err == nil {
		sc := session.NewSession(isp, ac, c)
		if sc != nil {
			factory.sessions = append(factory.sessions, sc)
		}
	} else {
		log.Error(err.Error())
	}

	winSize := ConfigYml.GetInt(isp + ".mt-window-size")
	if maxConns > 0 {
		factory.window = make(chan struct{}, winSize)
	} else {
		factory.window = make(chan struct{}, 16)
	}

	// 默认1W微妙即10毫秒生成一个token，也即tps最大200
	ev := 10 * time.Millisecond
	throughput := ConfigYml.GetInt(isp + ".throughput")
	if throughput > 0 {
		// 1s = 1000*1000 microsecond = 1000000 microsecond, Throughput 单位时TPS
		ev = time.Duration(1000000/throughput) * time.Microsecond
	}
	limit := rate.Every(ev)
	factory.limiter = rate.NewLimiter(limit, cap(factory.window))
	factory.startLruSortTicker()
	factory.RegCloseSessionsHooker()

	factories[ISP(isp).Int()] = factory
	return factory
}

// PeekSession 获取排序后在头部的会话（最近最少使用的会话）
func (f *SessionFactory) PeekSession() *session.Session {
	if !f.limiter.Allow() {
		return nil
	}

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

// StartCacheExpireTicker 过期数据定期检查器
func StartCacheExpireTicker(asyncHandler func([]any)) {
	go func() {
		d := ConfigYml.GetDuration("Cache.expire-check-duration")
		if d == 0 {
			d = time.Second
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			<-ticker.C
			// 1. 查询缓存的过期清晰与持久化
			cleanQueryCacheMap(asyncHandler)

			// 2. 请求响应缓存的清理
			cleanRequestIdCacheMap()

			// 3. 状态报告缓存清理
			cleanMsgIdCacheMap()
		}
	}()
}

func cleanQueryCacheMap(asyncHandler func([]any)) {
	expired := make([]int64, 0, 16)
	batch := make([]any, 0, 128)
	resultQueryCacheMap.Range(func(key, value any) bool {
		id := key.(int64)
		results := value.([]any)
		if len(results) == 0 {
			expired = append(expired, id)
		} else {
			d := ConfigYml.GetDuration("Cache.expire-time")
			if d == 0 {
				d = time.Minute
			}
			// 这里过期时间不需精准，我们只判断每个切片的第一个元素是否已经过程，如果过期，就整个切片删除
			r0 := results[0].(*session.Result)
			if r0.SendTime.Add(d).Before(time.Now()) {
				expired = append(expired, id)
				batch = append(batch, results...)
			}
		}
		// 如果过期处理器不为空，异步处理结果数据
		if asyncHandler != nil && len(batch) >= 64 {
			asyncHandler(batch)
			batch = make([]any, 0, 128)
		}
		return true
	})
	for _, key := range expired {
		key := key
		resultQueryCacheMap.Delete(key)
	}
	// 如果过期处理器不为空，异步处理结果数据
	if asyncHandler != nil && len(batch) > 0 {
		asyncHandler(batch)
	}
}

func cleanRequestIdCacheMap() {
	expiredKeys := make([]uint64, 0, 32)
	session.SequenceIdResultCacheMap.Range(func(key, value any) bool {
		d := ConfigYml.GetDuration("Cache.expire-time")
		if d == 0 {
			d = time.Minute
		}
		result := value.(*session.Result)
		if result.SendTime.Add(d).Before(time.Now()) {
			expiredKeys = append(expiredKeys, result.SequenceId)
		}
		return true
	})
	for _, key := range expiredKeys {
		key := key
		session.SequenceIdResultCacheMap.Delete(key)
	}
}

func cleanMsgIdCacheMap() {
	expiredKeys := make([]string, 0, 32)
	session.MsgIdResultCacheMap.Range(func(key, value any) bool {
		d := ConfigYml.GetDuration("Cache.expire-time")
		if d == 0 {
			d = time.Minute
		}
		result := value.(*session.Result)
		if result.SendTime.Add(d).Before(time.Now()) {
			expiredKeys = append(expiredKeys, result.MsgId)
		}
		return true
	})
	for _, key := range expiredKeys {
		key := key
		session.MsgIdResultCacheMap.Delete(key)
	}
}

func (f *SessionFactory) startLruSortTicker() {
	go func() {
		d := ConfigYml.GetDuration(f.srvName + ".tick-duration")
		if d == 0 {
			d = time.Second
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		cancel := event_manager.RegisterShutdownHookerAddChan(fmt.Sprintf("Stop_LruSortTicker_%p", f),
			func(args ...any) { ticker.Stop() },
		)
		for {
			<-ticker.C
			f.lruSort(cancel)
		}
	}()
}

func (f *SessionFactory) lruSort(cancel <-chan bool) {
	maxConns := ConfigYml.GetInt(f.srvName + ".max-conns")
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
		select {
		case <-cancel:
			log.Info("Receive signal and stop create new connects.")
			return
		default:
			f.newConnect()
		}
		//  使用固定间隔创建会话,避免瞬时创建太多
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
		sc := session.NewSession(f.srvName, f.authConf, c)
		if sc != nil {
			f.Lock()
			f.sessions = append(f.sessions, sc)
			f.Unlock()
		}
	} else {
		log.Error(err.Error())
	}
}

func (f *SessionFactory) RegCloseSessionsHooker() {
	event_manager.RegisterShutdownHooker(fmt.Sprintf("CloseSessionHooker_%p", f), func(args ...any) {
		close(f.window)
		for _, sc := range f.sessions {
			sc.Close()
			log.Warnf("session %p closed.", sc)
		}
	})
}

type ISP string

var ISPS = [3]string{session.CMPP, session.SGIP, session.SMGP}

func (i ISP) Int() int {
	switch i {
	case session.CMPP:
		return 0
	case session.SGIP:
		return 1
	case session.SMGP:
		return 2
	}
	log.Panicf("ISP \"%s\" not found!", i)
	return -1
}

func FindAuthConf(isp, clientId string) (ac *codec.AuthConf) {
	c := auth.Cache.FindByCid(isp, clientId)
	if c == nil {
		return nil
	}
	ac = &codec.AuthConf{}
	utils.StructCopy(c, ac)
	return
}
