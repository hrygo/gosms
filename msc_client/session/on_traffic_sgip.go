package session

import (
	"fmt"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/sgip"
)

func (s *Session) sendBySgip(phone string, message string, options ...codec.OptionFunc) (results []any) {
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)
	mts := sgip.NewSubmit(s.authConf, []string{phone}, message, options...)
	for _, mt := range mts {
		_, err := s.con.Write(mt.Encode())
		if err != nil {
			log.Error(err.Error())
			s.Close()
			return nil
		}
		mtt := mt.(*sgip.Submit)
		log.Debug(send, mtt.Log()...)

		r := Result{SendTime: time.Now()}
		r.SequenceId = mtt.Sequence2Uint64()
		r.Phone = phone
		results = append(results, &r)
		SequenceIdResultCacheMap.Store(r.SequenceId, &r)
	}
	return
}

func (s *Session) onTrafficSgip(cmd, seq uint32, buff []byte) {

	s.activeTime = time.Now()
	var receive = fmt.Sprintf("[%s] OnTraffic <<<", s.serverName)
	var send = fmt.Sprintf("[%s] OnTraffic >>>", s.serverName)

	switch sgip.CommandId(cmd) {
	case sgip.SGIP_BIND_RESP:
		resp := &sgip.BindRsp{}
		err := resp.Decode(seq, buff)
		log.Info(receive, resp.Log()...)
		if err != nil || resp.Status != sgip.Status(0) {
			log.Error(err.Error())
			s.Close()
		}
	case sgip.SGIP_UNBIND:
		term := &sgip.Unbind{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		resp := term.ToResponse(0)
		_, _ = s.con.Write(resp.Encode())
		log.Info(send, resp.Log()...)
		s.Close()
	case sgip.SGIP_UNBIND_RESP:
		term := &sgip.UnbindRsp{}
		_ = term.Decode(seq, buff)
		log.Info(receive, term.Log()...)
		s.Close()
	case sgip.SGIP_SUBMIT_RESP:
		sub := &sgip.SubmitRsp{}
		err := sub.Decode(seq, buff)
		log.Debug(receive, sub.Log()...)
		if err != nil {
			log.Error(err.Error())
			s.Close()
		}
		result, ok := SequenceIdResultCacheMap.Load(sub.Sequence2Uint64())
		if ok {
			mtr := result.(*Result)
			mtr.Result = uint32(sub.Status)
			mtr.MsgId = sub.Sequence2String()
			mtr.ResponseTime = time.Now()
			// 以msgId为Key存储到内存缓存
			MsgIdResultCacheMap.Store(mtr.MsgId, mtr)
		}
	}
}
