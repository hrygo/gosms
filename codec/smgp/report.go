package smgp

import (
	"fmt"
	"time"
)

type Report struct {
	id         []byte // 【10字节】状态报告对应原短消息的MsgID
	sub        string // 【3字节】取缺省值001
	dlvrd      string // 【3字节】取缺省值001
	submitDate string // 【10字节】短消息提交时间（格式：年年月月日日时时分分，例如010331200000）
	doneDate   string // 【10字节】短消息提交时间（格式：年年月月日日时时分分，例如010331200000）
	stat       string // 【7字节】短消息的最终状态
	err        string // 【3字节】短消息的最终状态
	txt        string // 【20字节】前3个字节，表示短消息长度（用ASCII码表示），后17个字节表示短消息的内容
}

const RptLen = 10 + 3 + 3 + 10 + 10 + 7 + 3 + 20 + len("id: sub: dlvrd: submit date: done date: stat: err: text:")

func NewReport(id []byte) *Report {
	report := &Report{id: id}
	report.sub = "001"
	report.dlvrd = "001"
	report.submitDate = time.Now().Format("0601021504")
	report.doneDate = time.Now().Add(time.Minute).Format("0601021504")
	report.txt = ""
	// 判断序号的时间戳部分
	switch time.Now().Unix() % 1000 {
	case 1:
		report.err = "001"
		report.stat = reportStatMap["001"]
	case 2:
		report.err = "002"
		report.stat = reportStatMap["002"]
	case 3:
		report.err = "003"
		report.stat = reportStatMap["003"]
	case 4:
		report.err = "004"
		report.stat = reportStatMap["004"]
	case 5:
		report.err = "005"
		report.stat = reportStatMap["005"]
	case 6:
		report.err = "006"
		report.stat = reportStatMap["006"]
	case 7:
		report.err = "007"
		report.stat = reportStatMap["007"]
	case 8:
		report.err = "008"
		report.stat = reportStatMap["008"]
	case 9:
		report.err = "009"
		report.stat = reportStatMap["009"]
	case 10:
		report.err = "010"
		report.stat = reportStatMap["010"]
	default:
		report.err = "000"
		report.stat = reportStatMap["000"]
	}
	return report
}

func (rt *Report) String() string {
	return fmt.Sprintf("id:%x sub:%s dlvrd:%s submit date:%s done date:%s stat:%s err:%s text:%x",
		rt.id, rt.sub, rt.dlvrd, rt.submitDate, rt.doneDate, rt.stat, rt.err, rt.txt)
}

func (rt *Report) Encode() []byte {
	data := make([]byte, RptLen)
	index := 0
	copy(data[index:index+3], "id:")
	index += 3
	copy(data[index:index+10], rt.id)
	index += 10
	str := rt.String() // 不含text的值
	start := 3 + 20    // "id:%x"
	copy(data[index:RptLen-20], str[start:])
	return data
}

func (rt *Report) Decode(frame []byte) error {
	index := 3 // skip "id:"
	rt.id = frame[index : index+10]
	index += 10

	index += 5 // skip " sub:"
	rt.sub = string(frame[index : index+3])
	index += 3

	index += 7 // skip " dlvrd:"
	rt.dlvrd = string(frame[index : index+3])
	index += 3

	index += 13 // skip " submit date:"
	rt.submitDate = string(frame[index : index+10])
	index += 10

	index += 11 // skip " done date:"
	rt.doneDate = string(frame[index : index+10])
	index += 10

	index += 6 // skip " stat:"
	rt.stat = string(frame[index : index+7])
	index += 7

	index += 5 // skip " err:"
	rt.err = string(frame[index : index+3])
	return nil
}

func (rt *Report) Id() []byte {
	return rt.id
}

func (rt *Report) Sub() string {
	return rt.sub
}

func (rt *Report) Dlvrd() string {
	return rt.dlvrd
}

func (rt *Report) SubmitDate() string {
	return rt.submitDate
}

func (rt *Report) DoneDate() string {
	return rt.doneDate
}

func (rt *Report) Stat() string {
	return rt.stat
}

var reportStatMap = map[string]string{
	"000": "DELIVRD", // 成功
	"001": "EXPIRED", // 用户不能通信
	"002": "EXPIRED", // 用户忙
	"003": "UNDELIV", // 终端无此部件号
	"004": "UNDELIV", // 非法用户
	"005": "UNDELIV", // 用户在黑名单内
	"006": "UNDELIV", // 系统错误
	"007": "EXPIRED", // 用户内存满
	"008": "UNDELIV", // 非信息终端
	"009": "UNDELIV", // 数据错误
	"010": "UNDELIV", // 数据丢失
	"999": "UNKNOWN", // 未知错误
}
