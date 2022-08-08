package server

import (
	"fmt"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosms/codec/cmpp"
)

var cmppActive TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_ACTIVE_TEST) != cmd {
		return true, gnet.None
	}

	sc := Session(c)
	if !sessionCheck(sc) {
		return false, gnet.Close
	}

	pdu := &cmpp.ActiveTest{}
	err := pdu.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(sc, buff)
		return false, gnet.Close
	}

	// 异步处理，避免阻塞 event-loop
	err = sc.Pool().Submit(func() {
		handleCmppActive(s, sc, pdu)
	})
	if err != nil {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC),
			FlatMapLog(sc.LogSession(), []log.Field{OpDropMessage.Field(), ErrorField(err), Packet2HexLogStr(pdu.Encode())})...)
		return false, gnet.Close
	}

	return false, gnet.None
}

func handleCmppActive(s *Server, sc *session, pdu *cmpp.ActiveTest) {
	// 【会话级别流控】采用通道控制消息收发速度,向通道发送信号
	sc.window <- struct{}{}
	defer func() { <-sc.window }()
	// 这里采用流量控制目的是防止客户端采用Active进行拒绝服务攻击

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)
	// 打印报文
	log.Debug(msg, FlatMapLog(sc.LogSession(), pdu.Log())...)

	// n. 发送响应
	resp := pdu.ToResponse(0)
	pack := resp.Encode()
	// 异步非阻塞
	err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		_ = c.Flush()
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Debug(msg, FlatMapLog(sc.LogSession(), resp.Log())...)
		// 更新会话
		sc.Lock()
		defer sc.Unlock()
		sc.lastUseTime = time.Now()
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_ACTIVE_TEST_RESP.OpLog(), SErrField(err.Error())})...)
	}
}
