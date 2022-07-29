package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec/smgp"
)

var smgpExit TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(smgp.SMGP_EXIT) != cmd {
		return true, gnet.None
	}

	sc := Session(c)
	if !sessionCheck(sc) {
		return false, gnet.Close
	}

	term := &smgp.Exit{}
	err := term.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(sc, buff)
		return false, gnet.Close
	}

	// 异步处理，避免阻塞 event-loop
	err = sc.Pool().Submit(func() {
		handleSmgpExit(s, sc, term)
	})
	if err != nil {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC),
			FlatMapLog(sc.LogSession(), []log.Field{OpDropMessage.Field(), ErrorField(err), Packet2HexLogStr(term.Encode())})...)
		return false, gnet.Close
	}

	return false, gnet.None
}

func handleSmgpExit(s *Server, sc *session, term *smgp.Exit) {
	// 【会话级别流控】采用通道控制消息收发速度,向通道发送信号
	sc.window <- struct{}{}
	defer func() { <-sc.window }()
	// 这里采用流量控制目的是防止客户端采用此消息进行拒绝服务攻击

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)
	// 打印报文
	log.Debug(msg, FlatMapLog(sc.LogSession(), term.Log())...)

	// n. 发送响应
	resp := term.ToResponse(0)
	pack := resp.Encode()
	// 异步非阻塞
	err := sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		_ = c.Flush()
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Debug(msg, FlatMapLog(sc.LogSession(), resp.Log())...)
		// 关闭连接
		_ = c.Close()
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{smgp.SMGP_EXIT.OpLog(), SErrField(err.Error())})...)
	}
}
