package session

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

func (s *Session) sendBySmgp(phone string, message string, options ...codec.OptionFunc) (results []any) {
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)
	mts := smgp.NewSubmit(s.authConf, []string{phone}, message, uint32(codec.B32Seq.NextVal()), options...)
	for _, mt := range mts {
		_, err := s.con.Write(mt.Encode())
		if err != nil {
			log.Error(err.Error())
			s.Close()
			return nil
		}
		mtt := mt.(*smgp.Submit)
		log.Debug(send, mtt.Log()...)

		r := Result{SendTime: time.Now()}
		r.SequenceId = uint64(mtt.SequenceId)
		r.Phone = phone
		results = append(results, &r)
		SequenceIdResultCacheMap.Store(r.SequenceId, &r)
	}
	return
}

func (s *Session) onTrafficSmgp(cmd, seq uint32, buff []byte) {
	s.activeTime = time.Now()
	var receive = fmt.Sprintf("[%s] OnTraffic <<<", s.serverName)
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)

	switch smgp.CommandId(cmd) {
	case smgp.SMGP_LOGIN_RESP:
		resp := &smgp.LoginRsp{Version: smgp.Version(s.authConf.Version)}
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
		sub := &smgp.SubmitRsp{Version: smgp.Version(s.authConf.Version)}
		err := sub.Decode(seq, buff)
		log.Debug(receive, sub.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
		result, ok := SequenceIdResultCacheMap.Load(sub.SequenceId)
		if ok {
			mtr := result.(*Result)
			mtr.Result = uint32(sub.Status())
			mtr.MsgId = hex.EncodeToString(sub.MsgId())
			mtr.ResponseTime = time.Now()
			// 已msgId为Key存储到内存缓存
			MsgIdResultCacheMap.Store(mtr.MsgId, mtr)
		}
	case smgp.SMGP_DELIVER:
		dly := &smgp.Delivery{Version: smgp.Version(s.authConf.Version)}
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
			key := hex.EncodeToString(rpt.Id())
			val, ok := MsgIdResultCacheMap.Load(key)
			if ok {
				mtr := val.(*Result)
				mtr.Report = rpt.Stat()
				mtr.ReportTime = time.Now()
			}
		}
	}
}
