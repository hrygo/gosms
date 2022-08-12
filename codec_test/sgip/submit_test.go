package sgip

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/sgip"
)

func TestSubmit(t *testing.T) {
	// Test New Log
	pdus := sgip.NewSubmit(ac, []string{"18600001111"}, Poem, codec.MtSpSubNo("010"))
	for _, pdu := range pdus {
		log.Info("submit", pdu.Log()...)
		submit := pdu.(*sgip.Submit)

		// Test Req Encode
		dt := submit.Encode()
		log.Infof("data: %x", dt)
		assert.True(t, uint32(len(dt)) == submit.PacketLength)

		// Test Req Decode
		err := submit.Decode(1, dt[12:])
		if err != nil {
			t.Error(err)
		}
		assert.True(t, err == nil)
		assert.True(t, submit.SequenceNumber[0] == 1)
		assert.True(t, submit.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
		log.Info("submit", submit.Log()...)

		// Test ToResponse Rsp.Log
		pdu2 := submit.ToResponse(0)
		log.Info("submitRsp", pdu2.Log()...)
		submitRsp := pdu2.(*sgip.SubmitRsp)

		// Test Resp Encode
		dt = submitRsp.Encode()
		log.Infof("data: %x", dt)
		assert.True(t, uint32(len(dt)) == submitRsp.PacketLength)

		// Test Resp Decode
		err = submitRsp.Decode(2, dt[12:])
		if err != nil {
			t.Error(err)
		}
		assert.True(t, err == nil)
		assert.True(t, submitRsp.SequenceNumber[0] == 2)
		assert.True(t, submitRsp.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
		log.Info("submitRsp", submitRsp.Log()...)
	}
}
