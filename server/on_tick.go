package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/my_errors"
	"github.com/hrygo/gosmsn/utils"
)

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnTick === %s", s.name, s.Address())
	log.Info(msg, s.LogCounterWithName()...)

	s.sessionPool.Range(func(key, value interface{}) bool {
		session, ok := value.(*session)
		if ok {
			// 关闭长时间未活动的连接
			pass := closeCheck(s, session)
			// 关闭心跳未正常响应的连接
			if pass {
				pass = activeTestNoneResponseCheck(s, session)
			}
			if pass {
				log.Info(msg, FlatMapLog(session.LogSession(), session.LogCounter())...)
			}
			// 发送心跳测试
			if pass && session.stat == StatLogin {
				activeTest(s, session)
			}
			// 发送模拟上行消息
			if pass && session.stat == StatLogin {
				mockDelivery(s, session)
			}
		}
		return true
	})
	return bs.ConfigYml.GetDuration("Server.TickDuration"), gnet.None
}

func activeTestNoneResponseCheck(s *Server, sc *session) bool {
	if sc.nAt > 2 {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		// 发送关闭指令
		sendTerminate(s, sc)
		log.Warn(msg, FlatMapLog(sc.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(my_errors.ErrorsNoneActiveTestResponse)})...)
		return false
	}
	return true
}

func closeCheck(s *Server, sc *session) bool {
	confTime := bs.ConfigYml.GetDuration("Server.ForceCloseConnTime")
	if sc.LastUseTime().Add(confTime).Before(time.Now()) {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		// 发送关闭指令
		sendTerminate(s, sc)
		log.Warn(msg, FlatMapLog(sc.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(fmt.Sprintf(my_errors.ErrorsNoEffectiveAction, confTime))})...)
		return false
	}
	return true
}

func activeTest(s *Server, sc *session) {
	_ = s.goPool.Submit(func() {
		// 如果使用时间在1分钟前则发送心跳
		if sc.lastUseTime.Add(time.Minute).Before(time.Now()) {
			msg := fmt.Sprintf("[%s] OnTick %s", s.name, SD)
			var active codec.RequestPdu
			var seq = uint32(codec.B32Seq.NextVal())
			switch s.name {
			case CMPP:
				active = cmpp.NewActiveTest(seq)
			case SGIP:
				// active =
			case SMGP:
				// active =
			}
			pack := active.Encode()
			err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
				sc.Lock()
				defer sc.Unlock()
				// 主动发起心跳不重置会话使用时间（也就是说即使服务端发送心跳，客户端长期不活跃一样会被断开连接）
				// sc.lastUseTime = time.Now()
				sc.nAt += 1
				return nil
			})
			if err == nil {
				log.Info(msg, FlatMapLog(sc.LogSession(), active.Log())...)
			} else {
				log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{SErrField(err.Error())})...)
			}
		}
	})
}

// 模拟发送随机上行短信
func mockDelivery(s *Server, sc *session) {
	// 开关检查
	open := bs.ConfigYml.GetBool("Mock.Delivery.Enable")
	if !open {
		return
	}
	cli := client.Cache.FindByCid(s.name, sc.clientId)
	if cli == nil {
		return
	}
	// 获取模拟消息内容
	contents := bs.ConfigYml.GetStringSlice("Mock.Delivery.Contents")
	if len(contents) == 0 {
		return
	}
	// 随机获取一条消息
	content := contents[utils.RandNum(0, len(contents))]
	csl := strings.Split(content, "_,_")
	if len(csl) < 2 {
		return
	}
	subNo := csl[0]
	text := csl[1]
	_ = s.goPool.Submit(func() {
		msg := fmt.Sprintf("[%s] OnTick %s", s.name, SD)
		var dly codec.RequestPdu
		var seq = uint32(codec.B32Seq.NextVal())
		switch s.name {
		case CMPP:
			dly = cmpp.NewDelivery(cli, "10011110000", text, cli.SmsDisplayNo+subNo, cli.ServiceId, seq)
		case SGIP:
			// dly =
		case SMGP:
			// dly =
		}
		pack := dly.Encode()
		err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
			sc.CounterAddDly()
			return nil
		})
		if err == nil {
			log.Info(msg, FlatMapLog(sc.LogSession(), dly.Log())...)
		} else {
			log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{SErrField(err.Error())})...)
		}
	})
}

func sendTerminate(s *Server, sc *session) {
	_ = s.goPool.Submit(func() {
		msg := fmt.Sprintf("[%s] OnTick %s", s.name, SD)
		var term codec.RequestPdu
		var seq = uint32(codec.B32Seq.NextVal())
		switch s.name {
		case CMPP:
			term = cmpp.NewTerminate(seq)
		case SGIP:
			// term =
		case SMGP:
			// term =
		}
		pack := term.Encode()
		err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
			_ = sc.Conn().Flush()
			// 给对方预留1秒钟响应连接关闭事件
			time.Sleep(time.Second)
			_ = sc.Conn().Close()
			return nil
		})
		if err == nil {
			log.Info(msg, FlatMapLog(sc.LogSession(), term.Log())...)
		} else {
			log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{SErrField(err.Error())})...)
		}
	})
}
