package server

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
)

func cmppOnTraffic(s *Server, c gnet.Conn) (action gnet.Action) {
	var msg = fmt.Sprintf("[%s] OnTraffic [%v<->%v]", s.name, c.RemoteAddr(), c.LocalAddr())

	// 检查缓存
	if c.InboundBuffered() < 12 {
		return
	}

	buff, _ := c.Peek(8)

	// 消息头检查
	pkl, cmd := codec.UnpackHead(buff)
	if pkl > cmpp.CMPP3_PACKET_MAX || pkl < 12 {
		log.Error(msg, ErrorField(codec.NewOpError(codec.ErrTotalLengthInvalid, fmt.Sprintf("cmppOnTraffic read pack len is %d", pkl))))
		return gnet.Close
	}

	scmd := cmpp.CommandId(cmd).String()
	if strings.HasSuffix(scmd, "UNKNOWN") {
		log.Error(msg, ErrorField(codec.NewOpError(codec.ErrCommandIdInvalid, fmt.Sprintf("cmppOnTraffic read command is  %x(%s)", cmd, scmd))))
		return gnet.Close
	}

	// 检查消息体长度
	if int(pkl) > c.InboundBuffered() {
		return
	}
	// 消息体通过长度检查后,跳过消息头的前8字节
	_, _ = c.Discard(8)

	// 读取消息体
	buff, _ = c.Peek(int(pkl - 8))
	defer func() { _, _ = c.Discard(int(pkl - 8)) }()
	log.Debug(msg, log.Uint32("pkl", pkl), log.String("cmd", scmd), log.String("packet", hex.EncodeToString(buff)))

	//  这里遵循开闭原则，采用责任链实现，对拓展开放，对修改关闭
	return ExecuteChain(handlers(), cmd, buff, c, s)
}

func handlers() []TrafficHandler {
	return []TrafficHandler{
		cmppConnect,
		cmppSubmit,
	}
}
