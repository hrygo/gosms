package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/utils"
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

	// 1. 包检查
	result, err := cmppSubmitPacketCheck(sc, mt)
	// 2. 消息签名处理、长短信处理等等
	// 3. 计费检查及计费
	// ...
	// 4. 模拟网关整体的处理耗时
	mockRandPrecessTime()
	// 5. 按比例模拟失败情况
	if utils.DiceCheck(bootstrap.ConfigYml.GetFloat64("Server.Mock.SuccessRate")) {
		result = uint32(cmpp.MtRsp8)
	}
	// ...
	// n. 发送响应
	resp := mt.ToResponse(result)
	pack := resp.Encode()
	// 异步非阻塞
	err = sc.conn.AsyncWrite(pack, func(c gnet.Conn) error {
		// 更新mt计数器
		sc.CounterAddMt()
		msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, SD)
		log.Debug(msg, FlatMapLog(sc.LogSession(16), resp.Log())...)
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_SUBMIT_RESP.Log(), SErrField(err.Error())})...)
	}
	// n+1. SMSC异步发送消息
	// ...
	// n+m. 模拟发送状态报告
	if result == uint32(cmpp.MtStatusOK) {
		rsp := resp.(*cmpp.SubmitResp)
		mockSendCmppReport(sc, mt, rsp.MsgId())
	}
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

func mockSendCmppReport(sc *session, sub *cmpp.Submit, msgId uint64) {
	// 按概率不返回状态报告
	if utils.DiceCheck(bootstrap.ConfigYml.GetFloat64("Server.Mock.SuccessRate")) {
		return
	}
	msg := fmt.Sprintf("[%s] OnTraffic %s", sc.serverName, SD)

	dly := sub.ToDeliveryReport(msgId)
	// 模拟状态报告发送前的耗时
	ms := bootstrap.ConfigYml.GetInt("Server.Mock.FixReportRespMs")
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	// 发送状态报告
	err := sc.conn.AsyncWrite(dly.Encode(), func(c gnet.Conn) error {
		log.Debug(msg, FlatMapLog(sc.LogSession(32), dly.Log())...)
		sc.CounterAddRpt()
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{cmpp.CMPP_DELIVER.Log(), SErrField(err.Error())})...)
	}
}
