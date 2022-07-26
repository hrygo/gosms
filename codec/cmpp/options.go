package cmpp

import (
	"time"
)

type Option func(mtOps *MtOptions)

func loadOptions(options ...Option) *MtOptions {
	opts := &MtOptions{
		RegisteredDel:   uint8(0xf),
		MsgLevel:        uint8(0xf),
		FeeUsertype:     uint8(0xf),
		FeeTerminalType: uint8(0xf),
	}
	for _, option := range options {
		option(opts)
	}
	return opts
}

type MtOptions struct {
	RegisteredDel   uint8
	MsgLevel        uint8
	FeeUsertype     uint8
	FeeTerminalType uint8
	ServiceId       string
	FeeTerminalId   string
	FeeType         string
	FeeCode         string
	ValidTime       string
	AtTime          string
	SrcId           string
	LinkID          string
}

// WithOptions 设置配置项
func WithOptions(opt *MtOptions) Option {
	return func(mtOps *MtOptions) {
		mtOps = opt
	}
}

// MtFeeTerminalType 被计费用户的号码类型，0：真实号码；1：伪码
func MtFeeTerminalType(t uint8) Option {
	if t != 0 && t != 1 {
		t = uint8(0xf)
	}
	return func(opts *MtOptions) {
		opts.FeeTerminalType = t
	}
}

// MtFeeUsertype 计费用户类型字段
// 0：对目的终端MSISDN计费；
// 1：对源终端MSISDN计费；
// 2：对SP计费;
// 3：表示本字段无效，对谁计费参见Fee_terminal_Id 字段。
func MtFeeUsertype(t uint8) Option {
	if t != 0 && t != 1 && t != 2 && t != 3 {
		t = uint8(0xf)
	}
	return func(opts *MtOptions) {
		opts.FeeUsertype = t
	}
}

// MtLinkID 点播业务使用的LinkID，非点播类业务的MT流程不使用该字段
func MtLinkID(s string) Option {
	return func(opts *MtOptions) {
		opts.LinkID = s
	}
}

// MtSrcId SP的服务代码或前缀为服务代码的长号码, 网关将该号码完整的填到SMPP协议Submit_SM消息相应的source_addr字段，该号码最终在用户手机上显示为短消息的主叫号码
func MtSrcId(s string) Option {
	return func(opts *MtOptions) {
		opts.SrcId = s
	}
}

// MtAtTime 定时发送时间，格式遵循SMPP3.3协议
func MtAtTime(t time.Time) Option {
	return func(opts *MtOptions) {
		s := t.Format("060102150405")
		opts.AtTime = s + "032+"
	}
}

// MtAtTimeStr 定时发送时间，格式:yyMMddHHmmss
func MtAtTimeStr(s string) Option {
	return func(opts *MtOptions) {
		if len(s) > 12 {
			s = s[:12]
		}
		opts.AtTime = s + "032+"
	}
}

// MtValidTime 存活有效期，格式遵循SMPP3.3协议
func MtValidTime(s string) Option {
	return func(opts *MtOptions) {
		opts.ValidTime = s
	}
}

// MtFeeCode 资费代码（以分为单位）
func MtFeeCode(s string) Option {
	return func(opts *MtOptions) {
		opts.FeeCode = s
	}
}

// MtFeeType 资费类别
// 01：对“计费用户号码”免费
// 02：对“计费用户号码”按条计信息费
// 03：对“计费用户号码”按包月收取信息费
// 04：对“计费用户号码”的信息费封顶
// 05：对“计费用户号码”的收费是由SP实现
func MtFeeType(s string) Option {
	if s != "01" && s != "02" && s != "03" && s != "04" && s != "05" {
		s = ""
	}
	return func(opts *MtOptions) {
		opts.FeeType = s
	}
}

// MtFeeTerminalId 计费号码与FeeTerminalType配合使用
func MtFeeTerminalId(id string) Option {
	return func(opts *MtOptions) {
		opts.FeeTerminalId = id
	}
}

// MtServiceId 业务标识，是数字、字母和符号的组合
func MtServiceId(id string) Option {
	return func(opts *MtOptions) {
		opts.ServiceId = id
	}
}

// MtRegisteredDel 是否需状态报告
func MtRegisteredDel(tf uint8) Option {
	if tf != 0 && tf != 1 {
		tf = uint8(0xf)
	}
	return func(opts *MtOptions) {
		opts.RegisteredDel = tf
	}
}

// MtMsgLevel 消息优先级
func MtMsgLevel(l uint8) Option {
	if l > 9 {
		l = uint8(0xf)
	}
	return func(opts *MtOptions) {
		opts.MsgLevel = l
	}
}
