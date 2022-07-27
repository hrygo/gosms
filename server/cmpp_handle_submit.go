package server

import (
	"fmt"
	"strings"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec/cmpp"
)

var cmppSubmit TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(cmpp.CMPP_SUBMIT) != cmd {
		// 转下一个handler处理
		return true, gnet.None
	}

	ses := Session(c)
	if !sessionCheck(ses) {
		return false, gnet.Close
	}

	var mt = &cmpp.Submit{Version: cmpp.Version(ses.ver)}
	err := mt.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(ses, buff)
		return false, gnet.Close
	}
	// 异步处理登录逻辑，避免阻塞 event-loop
	// 使用会话级别的 Pool, 这样不同会话之间的资源相对独立
	_ = ses.Pool().Submit(func() {
		handleCmppSubmit(s, ses, mt)
	})

	return false, gnet.None
}

func handleCmppSubmit(s *Server, sc *session, mt *cmpp.Submit) {
	// 【全局流控】采用通道控制消息收发速度,向通道发送信号
	s.window <- struct{}{}
	defer func() { <-s.window }()
	// 【会话级别流控】采用通道控制消息收发速度,向通道发送信号
	sc.window <- struct{}{}
	defer func() { <-sc.window }()

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)

	// 打印登录报文
	log.Debug(msg, FlatMapLog(sc.LogSession(32), mt.Log())...)

	var result uint32
	// 1. 包检查
	result, err := cmppSubmitPacketCheck(sc, mt)
	// 2. 消息签名处理、长短信处理等等
	// 3. 计费检查及计费
	// ...
	// n. 发送响应
	resp := mt.ToResponse(result)
	pack := resp.Encode()
	err = sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Debug(msg, FlatMapLog(sc.LogSession(16), resp.Log())...)
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_SUBMIT_RESP.Log(), SErrField(err.Error())})...)
	}
	// n+1. SMSC异步发送消息
	// ...
}

// 协议包检查，并根据检查情况给result赋值
func cmppSubmitPacketCheck(s *session, mt *cmpp.Submit) (result uint32, err error) {
	// 获取客户端信息
	cli := client.Cache.FindByCid(s.serverName, s.clientId)
	check := strings.HasPrefix(mt.SrcId(), cli.SmsDisplayNo)
	if check {
		check = mt.MsgLevel() < 10
	}
	// TODO more check
	return 0, nil
}
