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
	msg := fmt.Sprintf("[%s] OnTick @ %s", s.name, s.Address())
	log.Info(msg, log.Int(LogKeyActiveConns, s.engine.CountConnections()))

	s.sessions.Range(func(key, value interface{}) bool {
		session, ok := value.(*session)
		if ok {
			// 关闭长时间未活动的连接
			pass := closeCheck(s, session)

			// 发送心跳测试
			if pass && session.stat == StatLogin {
				// activeTest(s, session)
			}

			// 关闭心跳未正常响应的连接
			// if pass &&

		}
		return true
	})
	return bs.ConfigYml.GetDuration("Server.TickDuration"), gnet.None
}

func closeCheck(s *Server, ses *session) (pass bool) {
	pass = true
	// 5分钟未使用的连接
	if ses.LastUseTime().Add(5 * time.Second).Before(time.Now()) {
		msg := fmt.Sprintf("[%s] OnTick [%v<->%v]", s.name, ses.conn.RemoteAddr(), ses.conn.LocalAddr())
		conn := ses.Conn()
		_ = conn.Close()
		log.Warn(msg, ses.LogSid(), ConnectionClose.Field(), NoEffectiveActionTimeThresholdReached.Field())
		pass = false
	}
	return
}

func activeTest(s *Server, ses *session) {
	_ = s.pool.Submit(func() {
		msg := fmt.Sprintf("[%s] OnTick [%v<->%v]", s.name, ses.conn.RemoteAddr(), ses.conn.LocalAddr())
		var active codec.Packer
		var seq = uint32(codec.B32Seq.NextVal())
		switch s.name {
		case CMPP:
			active = &cmpp.ActiveTestReqPkt{SeqId: seq}
		case SGIP:
			// active =
		case SMGP:
			// active =
		}
		pack, err := active.Pack(seq)
		if err != nil {
			log.Error(msg, ses.LogSid(), ActiveTest.Field(), ErrorField(err))
			return
		}
		err = ses.conn.AsyncWrite(pack, nil)
		if err == nil {
			log.Info(msg, ses.LogSid(), ActiveTest.Field(), Packet2HexLogStr(pack))
		} else {
			log.Error(msg, ses.LogSid(), ActiveTest.Field(), ErrorField(err))
		}
	})
}
