package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/smgp"
	"github.com/hrygo/gosmsn/utils"
)

var smgpSubmit TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(smgp.SMGP_SUBMIT) != cmd {
		// 转下一个handler处理
		return true, gnet.None
	}

	sc := Session(c)
	if !sessionCheck(sc) {
		return false, gnet.Close
	}

	var mt = &smgp.Submit{Version: smgp.Version(sc.ver)}
	err := mt.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(sc, buff)
		return false, gnet.Close
	}

	pass := submitFlowControl(sc, mt, 1)
	if !pass {
		return false, gnet.None
	}

	// 异步处理，避免阻塞 event-loop
	// 使用会话级别的 GoPool, 这样不同会话之间的资源相对独立
	err = sc.Pool().Submit(func() {
		handleSmgpSubmit(s, sc, mt)
	})
	if err != nil {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC),
			FlatMapLog(sc.LogSession(), []log.Field{OpDropMessage.Field(), ErrorField(err), Packet2HexLogStr(buff)})...)
		return false, gnet.Close
	}

	return false, gnet.None
}

func handleSmgpSubmit(s *Server, sc *session, mt *smgp.Submit) {
	// 【会话级别流控】采用通道控制消息收发窗口,向通道发送信号
	sc.window <- struct{}{}
	defer func() { <-sc.window }()

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)

	// 打印登录报文
	log.Debug(msg, FlatMapLog(sc.LogSession(32), mt.Log())...)

	// 1. 包检查
	result, err := smgpSubmitPacketCheck(sc, mt)
	// 2. 消息签名处理、长短信处理等等
	// 3. 计费检查及计费
	// ...
	// 4. 模拟网关整体的处理耗时
	mockRandPrecessTime()
	// 5. 按比例模拟失败情况
	if utils.DiceCheck(bootstrap.ConfigYml.GetFloat64("Server.Mock.SuccessRate")) {
		result = 75
	}
	// ...
	// n. 发送响应
	resp, err := sendSubmitResponse(sc, mt, result)
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{smgp.SMGP_SUBMIT_RESP.OpLog(), SErrField(err.Error())})...)
		return
	}
	// n+1. SMSC异步发送消息
	// ...
	// n+m. 模拟发送状态报告
	if result == 0 {
		rsp := resp.(*smgp.SubmitRsp)
		mockSendSmgpReport(sc, mt, rsp.MsgId())
	}
}

// 协议包检查，并根据检查情况给result赋值
func smgpSubmitPacketCheck(s *session, mt *smgp.Submit) (result uint32, err error) {
	// 获取客户端信息
	cli := auth.Cache.FindByCid(s.serverName, s.clientId)
	check := strings.HasPrefix(mt.SrcTermID(), cli.SmsDisplayNo)
	if check {
		check = mt.Priority() < 4
	}
	// do more check
	// ...
	return 0, nil
}

func mockSendSmgpReport(sc *session, sub *smgp.Submit, msgId []byte) {
	// 按概率不返回状态报告
	if utils.DiceCheck(bootstrap.ConfigYml.GetFloat64("Server.Mock.SuccessRate")) {
		return
	}
	msg := fmt.Sprintf("[%s] OnTraffic %s", sc.serverName, SD)

	cli := auth.Cache.FindByCid(sc.serverName, sc.clientId)
	dly := smgp.NewDeliveryReport(cli, sub, uint32(codec.B32Seq.NextVal()), msgId)
	// 模拟状态报告发送前的耗时
	ms := bootstrap.ConfigYml.GetInt("Server.Mock.FixReportRespMs")
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	// 发送状态报告
	err := sc.conn.AsyncWrite(dly.Encode(), func(c gnet.Conn) error {
		_ = c.Flush()
		log.Debug(msg, FlatMapLog(sc.LogSession(32), dly.Log())...)
		sc.CounterAddRpt()
		return nil
	})
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{smgp.SMGP_DELIVER.OpLog(), SErrField(err.Error())})...)
	}
}
