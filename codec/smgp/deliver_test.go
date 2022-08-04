package smgp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosmsn/codec"
)

func TestDeliver_Decode(t *testing.T) {
	dlv := NewDeliver(cli, "123", "95535", "TD:123456", uint32(codec.B32Seq.NextVal()))
	t.Logf("dlv: %s", dlv)
	testDeliver(t, dlv)
}

func TestDeliver_ReportDecode(t *testing.T) {
	mts := NewSubmit(cli, []string{"17011113333"}, "hello world，世界", uint32(codec.B32Seq.NextVal()), MtOptions{})
	mt := mts[0]
	msp := mt.ToResponse(0).(*SubmitRsp)
	tm := mt.(*Submit)
	rpt := NewDeliveryReport(cli, tm, uint32(codec.B32Seq.NextVal()), msp.msgId)
	t.Logf("dlv: %s", rpt)
	testDeliver(t, rpt)
}

func testDeliver(t *testing.T, pdu codec.RequestPdu) {
	dlv := pdu.(*Delivery)
	resp := dlv.ToResponse(0).(*DeliverRsp)
	t.Logf("resp: %s", resp)

	// 测试Deliver Encode
	dt := dlv.Encode()
	assert.True(t, int(dlv.PacketLength) == len(dt))
	t.Logf("dlv_encode: %x", dt)
	// 测试Deliver Decode
	dlvDec := &Delivery{}
	err := dlvDec.Decode(dlv.SequenceId, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, dlvDec.MessageHeader.SequenceId == dlv.MessageHeader.SequenceId)
	t.Logf("dlv_decode: %s", dlvDec)

	// 测试DeliverResp Encode
	dt = resp.Encode()
	assert.True(t, int(resp.PacketLength) == len(dt))
	t.Logf("resp_encode: %x", dt)
	// 测试Deliver Decode
	respDec := &DeliverRsp{}
	err = respDec.Decode(dlv.SequenceId, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, respDec.MessageHeader.SequenceId == respDec.MessageHeader.SequenceId)
	t.Logf("resp_decode: %s", dlvDec)
}
