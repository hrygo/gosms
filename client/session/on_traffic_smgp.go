package session

import (
	"fmt"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec/smgp"
)

func (s *Session) onTrafficSmgp(cmd, seq uint32, buff []byte) {

	s.activeTime = time.Now()
	var receive = fmt.Sprintf("[%s] OnTraffic <<<", s.serverName)
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)

	switch smgp.CommandId(cmd) {
	case smgp.SMGP_LOGIN_RESP:
		resp := &smgp.LoginRsp{Version: smgp.Version(s.cli.Version)}
		err := resp.Decode(seq, buff)
		log.Info(receive, resp.Log()...)
		if err != nil || resp.Status() != smgp.Status(0) {
			log.Error(err.Error())
			s.Close()
		}
	case smgp.SMGP_ACTIVE_TEST:
		act := &smgp.ActiveTest{}
		_ = act.Decode(seq, buff)
		log.Info(receive, act.Log()...)
		resp := act.ToResponse(0)
		_, err := s.con.Write(resp.Encode())
		log.Info(send, resp.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
	case smgp.SMGP_ACTIVE_TEST_RESP:
		act := &smgp.ActiveTestRsp{}
		_ = act.Decode(seq, buff)
		log.Info(receive, act.Log()...)
	case smgp.SMGP_EXIT:
		term := &smgp.Exit{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		resp := term.ToResponse(0)
		_, _ = s.con.Write(resp.Encode())
		log.Info(send, resp.Log()...)
		s.Close()
	case smgp.SMGP_EXIT_RESP:
		term := &smgp.ExitRsp{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		s.Close()
	case smgp.SMGP_SUBMIT_RESP:
		sub := &smgp.SubmitRsp{Version: smgp.Version(s.cli.Version)}
		err := sub.Decode(seq, buff)
		log.Debug(receive, sub.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
	case smgp.SMGP_DELIVER:
		dly := &smgp.Delivery{Version: smgp.Version(s.cli.Version)}
		err := dly.Decode(seq, buff)
		log.Debug(receive, dly.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
			return
		}
		resp := dly.ToResponse(0)
		_, err = s.con.Write(resp.Encode())
		log.Debug(send, resp.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
	}
}
