package server

import (
	"fmt"
	"strings"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosms/auth"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/msc_server"
	"github.com/hrygo/gosms/utils"
)

var sgipSubmit TrafficHandler = func(cmd, seq uint32, buff []byte, c gnet.Conn, s *Server) (next bool, action gnet.Action) {
	if uint32(sgip.SGIP_SUBMIT) != cmd {
		// 转下一个handler处理
		return true, gnet.None
	}

	sc := Session(c)
	if !sessionCheck(sc) {
		return false, gnet.Close
	}

	var mt = &sgip.Submit{}
	err := mt.Decode(seq, buff)
	if err != nil {
		decodeErrorLog(sc, buff)
		return false, gnet.Close
	}

	pass := submitFlowControl(sc, mt, 33)
	if !pass {
		return false, gnet.None
	}

	// 异步处理，避免阻塞 event-loop
	// 使用会话级别的 GoPool, 这样不同会话之间的资源相对独立
	err = sc.Pool().Submit(func() {
		handleSgipSubmit(s, sc, mt)
	})
	if err != nil {
		log.Error(fmt.Sprintf("[%s] OnTraffic %s", sc.ServerName(), RC),
			FlatMapLog(sc.LogSession(), []log.Field{OpDropMessage.Field(), ErrorField(err), Packet2HexLogStr(buff)})...)
		return false, gnet.Close
	}

	return false, gnet.None
}

func handleSgipSubmit(s *Server, sc *session, mt *sgip.Submit) {
	// 【会话级别流控】采用通道控制消息收发窗口,向通道发送信号
	sc.window <- struct{}{}
	defer func() { <-sc.window }()

	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)

	// 打印报文
	log.Debug(msg, FlatMapLog(sc.LogSession(32), mt.Log())...)

	// 1. 包检查
	result, _ := sgipSubmitPacketCheck(sc, mt)
	// 2. 消息签名处理、长短信处理等等
	// 3. 计费检查及计费
	// ...
	// 4. 模拟网关整体的处理耗时
	mockRandPrecessTime()
	// 5. 按比例模拟失败情况
	if utils.DiceCheck(msc.ConfigYml.GetFloat64("Server.Mock.SuccessRate")) {
		result = 33
	}
	// ...
	// n. 发送响应
	_, err := sendSubmitResponse(sc, mt, result)
	if err != nil {
		log.Error(msg, FlatMapLog(sc.LogSession(), []log.Field{sgip.SGIP_SUBMIT_RESP.OpLog(), SErrField(err.Error())})...)
		return
	}
	// n+1. SMSC异步发送消息
	// ...
	// n+m. 模拟发送状态报告
	// TODO 建立链接发送状态报告
}

// 协议包检查，并根据检查情况给result赋值
func sgipSubmitPacketCheck(s *session, mt *sgip.Submit) (result uint32, err error) {
	// 获取客户端信息
	cli := auth.Cache.FindByCid(s.serverName, s.clientId)
	check := strings.HasPrefix(mt.SPNumber, cli.SmsDisplayNo)
	if check {
		check = mt.Priority < 10
	}
	// do more check
	// ...
	return 0, nil
}
