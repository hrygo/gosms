package server

import (
	"fmt"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/my_errors"
)

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnTick === %s", s.name, s.Address())
	log.Info(msg, s.LogCounterWithName()...)

	s.sessions.Range(func(key, value interface{}) bool {
		session, ok := value.(*session)
		if ok {
			// 关闭长时间未活动的连接
			pass := closeCheck(s, session)
			// 关闭心跳未正常响应的连接
			if pass {
				pass = activeTestNoneResponseCheck(s, session)
			}
			// 发送心跳测试
			if pass && session.stat == StatLogin {
				activeTest(s, session)
				log.Info(msg, FlatMapLog(session.LogSession(), session.LogCounter())...)
			}
		}
		return true
	})
	return bs.ConfigYml.GetDuration("Server.TickDuration"), gnet.None
}

func activeTestNoneResponseCheck(s *Server, session *session) bool {
	if session.nAt > 2 {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		conn := session.Conn()
		_ = conn.Close()
		log.Warn(msg, FlatMapLog(session.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(my_errors.ErrorsNoneActiveTestResponse)})...)
		return false
	}
	return true
}

func closeCheck(s *Server, ses *session) bool {
	confTime := bs.ConfigYml.GetDuration("Server.ForceCloseConnTime")
	if ses.LastUseTime().Add(confTime).Before(time.Now()) {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		conn := ses.Conn()
		_ = conn.Close()
		log.Warn(msg, FlatMapLog(ses.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(fmt.Sprintf(my_errors.ErrorsNoEffectiveAction, confTime))})...)
		return false
	}
	return true
}

func activeTest(s *Server, ses *session) {
	_ = s.pool.Submit(func() {
		// 如果适用时间在1分钟前则发送心跳
		if ses.lastUseTime.Add(time.Minute).Before(time.Now()) {
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
			err := ses.conn.AsyncWrite(pack, func(c gnet.Conn) error {
				ses.Lock()
				ses.Unlock()
				ses.lastUseTime = time.Now()
				ses.nAt += 1
				return nil
			})
			if err == nil {
				log.Info(msg, FlatMapLog(ses.LogSession(), []log.Field{OpActiveTest.Field()}, active.Log())...)
			} else {
				log.Error(msg, FlatMapLog(ses.LogSession(), []log.Field{OpActiveTest.Field(), SErrField(err.Error())})...)
			}
		}
	})
}
