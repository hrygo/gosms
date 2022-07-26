package server

import (
	"fmt"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
)

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnTick === %s", s.name, s.Address())
	log.Info(msg, CCWW(s)...)

	s.sessions.Range(func(key, value interface{}) bool {
		session, ok := value.(*session)
		if ok {
			// 关闭长时间未活动的连接
			pass := closeCheck(s, session)

			// 发送心跳测试
			if pass && session.stat == StatLogin {
				activeTest(s, session)
			}

			// 关闭心跳未正常响应的连接
			if pass {
				activeTestNoneResponseCheck(s, session)
			}
		}
		return true
	})
	return bs.ConfigYml.GetDuration("Server.TickDuration"), gnet.None
}

func activeTestNoneResponseCheck(s *Server, session *session) {
	if session.nAt > 2 {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		conn := session.Conn()
		_ = conn.Close()
		log.Warn(msg, JoinLog(SSR(session, conn.RemoteAddr()),
			ConnectionClose.Field(), NoneResponseActiveTestCountThresholdReached.Field())...)
	}
}

func closeCheck(s *Server, ses *session) (pass bool) {
	pass = true
	confTime := bs.ConfigYml.GetDuration("Server.ForceCloseConnTime")
	if ses.LastUseTime().Add(confTime).Before(time.Now()) {
		msg := fmt.Sprintf("[%s] OnTick ===", s.name)
		conn := ses.Conn()
		_ = conn.Close()
		log.Warn(msg, JoinLog(SSR(ses, conn.RemoteAddr()), ConnectionClose.Field(), NoEffectiveActionTimeThresholdReached.Field())...)
		pass = false
	}
	return
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
				log.Info(msg, ses.LogSid(), ActiveTest.Field(), Packet2HexLogStr(pack))
			} else {
				log.Error(msg, ses.LogSid(), ActiveTest.Field(), ErrorField(err))
			}
		}
	})
}
