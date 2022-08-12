package sgip

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

func TestDeliver(t *testing.T) {
	// Test New Log
	pdu := sgip.NewDeliver(ac, "18600001111", Poem, "01")
	log.Info("deliver", pdu.Log()...)
	deliver := pdu.(*sgip.Deliver)

	// Test Req Encode
	dt := deliver.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == deliver.PacketLength)

	// Test Req Decode
	err := deliver.Decode(1, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, deliver.SequenceNumber[0] == 1)
	assert.True(t, deliver.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("deliver", deliver.Log()...)

	// Test ToResponse Rsp.Log
	pdu2 := deliver.ToResponse(0)
	log.Info("dlvRsp", pdu2.Log()...)
	dlvRsp := pdu2.(*sgip.DeliverRsp)

	// Test Resp Encode
	dt = dlvRsp.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == dlvRsp.PacketLength)

	// Test Resp Decode
	err = dlvRsp.Decode(2, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, dlvRsp.SequenceNumber[0] == 2)
	assert.True(t, dlvRsp.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("dlvRsp", dlvRsp.Log()...)

}
