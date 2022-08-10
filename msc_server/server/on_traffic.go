package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/cmpp"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/codec/smgp"
	"github.com/hrygo/gosms/msc_server"
	"github.com/hrygo/gosms/utils"
)

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	cmd, seq, buff, action, ok := DecodeAndCheckHeader(s, c)
	if ok {
		switch s.name {
		case CMPP:
			return ExecuteChain(CmppHandlers(), cmd, seq, buff, c, s)
		case SGIP:
			return ExecuteChain(SgipHandlers(), cmd, seq, buff, c, s)
		case SMGP:
			return ExecuteChain(SmgpHandlers(), cmd, seq, buff, c, s)
		}
	}
	return action
}

// TrafficHandler Don't Read From gnet.Conn.
// cmd: command id ;
// seq: request sequence ;
// buff: packet body ;
// next: true continue next handler, false return the action
type TrafficHandler func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action)

// ExecuteChain all cmppHandlers
func ExecuteChain(handlers []TrafficHandler, cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (action gnet.Action) {
	for _, handler := range handlers {
		handler := handler
		next, action := handler(cmd, seq, buff, c, s)
		if next {
			continue
		} else {
			return action
		}
	}
	return action
}

func DecodeAndCheckHeader(s *Server, c gnet.Conn) (cmd uint32, seq uint32, buff []byte, action gnet.Action, pass bool) {
	// 检查缓存
	if c.InboundBuffered() < 12 {
		return 0, 0, nil, gnet.None, false
	}
	buff, _ = c.Peek(12)

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)
	sc := Session(c)

	// 消息头检查
	pkl, cmd, seq := codec.UnpackHead(buff)
	if pkl > codec.PacketMax || pkl < codec.HeadLen {
		log.Error(msg, FlatMapLog(sc.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(fmt.Sprintf(msc.ErrorsIllegalPacketLength, pkl)), Packet2HexLogStr(buff)})...)
		return 0, 0, nil, gnet.Close, false
	}

	// 命令检查
	var op codec.Operation
	switch s.name {
	case CMPP:
		op = cmpp.CommandId(cmd)
	case SGIP:
		op = sgip.CommandId(cmd)
	case SMGP:
		op = smgp.CommandId(cmd)
	}
	if strings.HasSuffix(op.String(), "UNKNOWN") {
		log.Error(msg, FlatMapLog(sc.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(fmt.Sprintf(msc.ErrorsIllegalCommand, cmd))})...)
		return 0, 0, nil, gnet.Close, false
	}

	// 检查消息体长度
	if int(pkl) > c.InboundBuffered() {
		return 0, 0, nil, gnet.None, false
	}
	// 消息体通过长度检查后,跳过消息头的前8字节
	_, _ = c.Discard(12)

	// 读取消息体
	buff, _ = c.Peek(int(pkl - 12))
	_, _ = c.Discard(int(pkl - 12))
	// buff returned by Peek() is not allowed to be passed to a new goroutine, as this []byte will be reused within event-loop.
	// If you have to use buf in a new goroutine, then you need to make a copy of buf and pass this copy to that new goroutine.
	newBuff := make([]byte, len(buff))
	copy(newBuff, buff)
	log.Debug(msg, FlatMapLog(sc.LogSession(), []log.Field{op.OpLog(), Packet2HexLogStr(newBuff)})...)

	return cmd, seq, newBuff, gnet.None, true
}

// 流量控制检查，如果检查不通过返回客户端 errCode 指定的错误码
func submitFlowControl(sc *session, mt codec.RequestPdu, errCode uint32) bool {
	// 限速检查
	ok := sc.limiter.Allow()
	if !ok {
		//  流量控制错误响应结果异步发送
		_ = sc.Pool().Submit(func() {
			_, _ = sendSubmitResponse(sc, mt, errCode)
		})
		msg := fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC)
		log.Warn(msg, FlatMapLog(sc.LogSession(), []log.Field{SErrField(msc.ErrorsSubmitFlowControl)})...)
	}
	return ok
}

// 发送下行短信响应
func sendSubmitResponse(sc *session, pdu codec.RequestPdu, result uint32) (codec.Pdu, error) {
	msg := fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), SD)
	resp := pdu.ToResponse(result)
	pack := resp.Encode()
	// 异步非阻塞
	err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		_ = c.Flush()
		if result == 0 {
			sc.CounterAddMt()
		}
		log.Debug(msg, FlatMapLog(sc.LogSession(16), resp.Log())...)
		return nil
	})
	return resp, err
}

func sessionCheck(s *session) bool {
	if s.stat != StatLogin {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", s.ServerName(), RC),
			FlatMapLog(s.LogSession(), []log.Field{OpConnectionClose.Field(), SErrField(msc.ErrorsDecodePacketBody)})...)
		return false
	}
	return true
}

func decodeErrorLog(s *session, buff []byte) {
	log.Error(fmt.Sprintf("[%s] OnTraffic %s", s.ServerName(), RC),
		FlatMapLog(s.LogSession(), []log.Field{OpConnectionClose.Field(), SErrField(msc.ErrorsDecodePacketBody), Packet2HexLogStr(buff)})...)
}

func mockRandPrecessTime() {
	min := msc.ConfigYml.GetInt("Server.Mock.MinSubmitRespMs")
	max := msc.ConfigYml.GetInt("Server.Mock.MaxSubmitRespMs")
	rt := time.Duration(utils.RandNum(min, max))
	time.Sleep(rt * time.Millisecond)
}
