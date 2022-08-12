package sgip

import (
	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/utils"
	"github.com/hrygo/gosms/utils/snowflake"
)

var conf = []byte(`
{
    "isp": "sgip",
		"loginName": "3037196688",
    "clientId": "3037196688",
    "sharedSecret": "shared secret",
    "version": 18,
    "needReport": 1,
    "smsDisplayNo": "95566",
    "serviceId": "myService",
    "DefaultMsgLevel": 3,
    "feeUserType": 2,
    "FeeTerminalType": 0,
    "feeTerminalId": "",
    "feeType": "05",
    "feeCode": "free",
    "fixedFee": "",
    "LinkId": "",
    "mtValidDuration": 7200000000000,
    "maxConns": 4,
    "mtWindowSize": 16,
    "throughput": 1000
}`)

var ac *codec.AuthConf

func init() {
	ac = codec.Unmarshal(conf)

	codec.B32Seq = utils.NewCycleSequence(1, 7)
	codec.B64Seq = snowflake.NewSnowflake(7, 110)
	codec.BcdSeq = utils.NewBcdSequence("010101")
	sgip.Sequencer = &sgip.SequenceNumber{Node: 3037196688}
}
