package session

import (
	"fmt"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec/cmpp"
)

func (s *Session) onTrafficCmpp(cmd, seq uint32, buff []byte) {
	s.activeTime = time.Now()
	var receive = fmt.Sprintf("[%s] OnTraffic <<<", s.serverName)
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)

	switch cmpp.CommandId(cmd) {
	case cmpp.CMPP_CONNECT_RESP:
		resp := &cmpp.ConnectResp{Version: cmpp.Version(s.cli.Version)}
		err := resp.Decode(seq, buff)
		log.Info(receive, resp.Log()...)
		if err != nil || resp.Status() != cmpp.ConnStatusOK {
			log.Error(err.Error())
			s.Close()
		}
	case cmpp.CMPP_ACTIVE_TEST:
		act := &cmpp.ActiveTest{}
		_ = act.Decode(seq, buff)
		log.Info(receive, act.Log()...)
		resp := act.ToResponse(0)
		_, err := s.con.Write(resp.Encode())
		log.Info(send, resp.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
	case cmpp.CMPP_ACTIVE_TEST_RESP:
		act := &cmpp.ActiveTestRsp{}
		_ = act.Decode(seq, buff)
		log.Info(receive, act.Log()...)
	case cmpp.CMPP_TERMINATE:
		term := &cmpp.Terminate{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		resp := term.ToResponse(0)
		_, _ = s.con.Write(resp.Encode())
		log.Info(send, resp.Log()...)
		s.Close()
	case cmpp.CMPP_TERMINATE_RESP:
		term := &cmpp.TerminateRsp{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		s.Close()
	case cmpp.CMPP_SUBMIT_RESP:
		sub := &cmpp.SubmitRsp{Version: cmpp.Version(s.cli.Version)}
		err := sub.Decode(seq, buff)
		log.Debug(receive, sub.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
	case cmpp.CMPP_DELIVER:
		dly := &cmpp.Delivery{Version: cmpp.Version(s.cli.Version)}
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
