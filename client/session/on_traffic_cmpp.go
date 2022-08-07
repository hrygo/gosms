package session

import (
	"fmt"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/utils"
)

func (s *Session) sendByCmpp(phone string, message string) (results []any) {
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)
	mts := cmpp.NewSubmit(s.cli, []string{phone}, message, uint32(codec.B32Seq.NextVal()))
	for _, mt := range mts {
		_, err := s.con.Write(mt.Encode())
		if err != nil {
			log.Error(err.Error())
			s.Close()
			return nil
		}
		mtt := mt.(*cmpp.Submit)
		log.Debug(send, mtt.Log()...)

		r := Result{SendTime: time.Now()}
		r.RequestId = mtt.SequenceId
		r.Phone = phone
		results = append(results, &r)
		RequestIdResultCacheMap.Store(r.RequestId, &r)
	}
	return
}

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
		result, ok := RequestIdResultCacheMap.Load(sub.SequenceId)
		if ok {
			mtr := result.(*Result)
			mtr.Result = sub.Result()
			mtr.MsgId = utils.Uint64HexString(sub.MsgId())
			mtr.ResponseTime = time.Now()
			// 已msgId为Key存储到内存缓存
			MsgIdResultCacheMap.Store(mtr.MsgId, mtr)
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
		// 是状态报告
		if dly.IsReport() {
			rpt := dly.Report()
			if rpt == nil {
				return
			}
			key := utils.Uint64HexString(rpt.MsgId())
			val, ok := MsgIdResultCacheMap.Load(key)
			if ok {
				mtr := val.(*Result)
				mtr.Report = rpt.Stat()
				mtr.ReportTime = time.Now()
			}
		}
	}
}
