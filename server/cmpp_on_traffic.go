package server

import (
	"fmt"
	"strings"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/my_errors"
)

func cmppOnTraffic(s *Server, c gnet.Conn) (action gnet.Action) {
	var msg = fmt.Sprintf("[%s] OnTraffic %s", s.name, RC)
	session := Session(c)
	// 检查缓存
	if c.InboundBuffered() < 12 {
		return
	}

	buff, _ := c.Peek(12)

	// 消息头检查
	pkl, cmd, seq := codec.UnpackHead(buff)
	if pkl > cmpp.PacketMax || pkl < 12 {
		log.Error(msg, FlatMapLog(session.LogSession(),
			[]log.Field{OpConnectionClose.Field(), SErrField(fmt.Sprintf(my_errors.ErrorsIllegalPacketLength, pkl)), Packet2HexLogStr(buff)})...)
		return gnet.Close
	}

	op := cmpp.CommandId(cmd)
	if strings.HasSuffix(op.String(), "UNKNOWN") {
		log.Error(msg, FlatMapLog(session.LogSession(),
			[]log.Field{SErrField(fmt.Sprintf(my_errors.ErrorsIllegalCommand, cmd))})...)
		return gnet.Close
	}

	// 检查消息体长度
	if int(pkl) > c.InboundBuffered() {
		return
	}
	// 消息体通过长度检查后,跳过消息头的前8字节
	_, _ = c.Discard(12)

	// 读取消息体
	buff, _ = c.Peek(int(pkl - 12))
	defer func() { _, _ = c.Discard(int(pkl - 12)) }()
	log.Debug(msg, FlatMapLog(session.LogSession(), []log.Field{op.Log(), Packet2HexLogStr(buff)})...)

	//  这里遵循开闭原则，采用责任链实现，对拓展开放，对修改关闭
	return ExecuteChain(handlers(), cmd, seq, buff, c, s)
}

func handlers() []TrafficHandler {
	return []TrafficHandler{
		cmppConnect,
		cmppSubmit,
	}
}
