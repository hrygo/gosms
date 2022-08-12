package smgp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

func TestDeliver_Decode(t *testing.T) {
	dlv := smgp.NewDeliver(ac, "123", "95535", "TD:123456", uint32(codec.B32Seq.NextVal()))
	t.Logf("dlv: %s", dlv)
	testDeliver(t, dlv)
}

func TestDeliver_ReportDecode(t *testing.T) {
	mts := smgp.NewSubmit(ac, []string{"17011113333"}, "hello world，世界", uint32(codec.B32Seq.NextVal()))
	mt := mts[0]
	msp := mt.ToResponse(0).(*smgp.SubmitRsp)
	tm := mt.(*smgp.Submit)
	rpt := smgp.NewDeliveryReport(ac, tm, uint32(codec.B32Seq.NextVal()), msp.MsgId())
	t.Logf("dlv: %s", rpt)
	testDeliver(t, rpt)
}

func testDeliver(t *testing.T, pdu codec.RequestPdu) {
	dlv := pdu.(*smgp.Delivery)
	resp := dlv.ToResponse(0).(*smgp.DeliverRsp)
	t.Logf("resp: %s", resp)

	// 测试Deliver Encode
	dt := dlv.Encode()
	assert.True(t, int(dlv.PacketLength) == len(dt))
	t.Logf("dlv_encode: %x", dt)
	// 测试Deliver Decode
	dlvDec := &smgp.Delivery{}
	err := dlvDec.Decode(dlv.SequenceId, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, dlvDec.MessageHeader.SequenceId == dlv.MessageHeader.SequenceId)
	t.Logf("dlv_decode: %s", dlvDec)

	// 测试DeliverResp Encode
	dt = resp.Encode()
	assert.True(t, int(resp.PacketLength) == len(dt))
	t.Logf("resp_encode: %x", dt)
	// 测试Deliver Decode
	respDec := &smgp.DeliverRsp{}
	err = respDec.Decode(dlv.SequenceId, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, respDec.MessageHeader.SequenceId == respDec.MessageHeader.SequenceId)
	t.Logf("resp_decode: %s", dlvDec)
}
