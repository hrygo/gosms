package smgp

import (
	"time"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/utils"
)

type MtOptions struct {
	NeedReport    byte          // SP是否要求返回状态报告
	Priority      byte          // 短消息发送优先级,0-3
	ServiceID     string        // 业务代码
	AtTime        time.Time     // 短消息定时发送时间
	ValidDuration time.Duration // 短消息有效时长
	SrcTermID     string        // 会拼接到配置文件的sms-display-no后面
}

func (s *Submit) SetOptions(cli *auth.Client, options MtOptions) {
	s.needReport = byte(cli.NeedReport)
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.NeedReport != 0 {
		s.needReport = options.NeedReport
	}

	s.priority = cli.DefaultMsgLevel
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.Priority != 0 {
		s.priority = options.Priority
	}

	s.serviceID = cli.ServiceId
	if options.ServiceID != "" {
		s.serviceID = options.ServiceID
	}

	if options.AtTime.Year() != 1 {
		s.atTime = utils.FormatTime(options.AtTime)
	} else {
		s.atTime = utils.FormatTime(time.Now())
	}

	vt := time.Now()
	if options.ValidDuration != 0 {
		vt.Add(options.ValidDuration)
	} else {
		vt.Add(cli.MtValidDuration)
	}
	s.validTime = utils.FormatTime(vt)

	s.srcTermID = cli.SmsDisplayNo
	if options.SrcTermID != "" {
		s.srcTermID += options.SrcTermID
	}
}
