package sgip

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

func TestUnbind(t *testing.T) {
	// Test New Log
	unbind := sgip.NewUnbind()
	log.Info("unbind", unbind.Log()...)

	// Test Req Encode
	dt := unbind.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == unbind.PacketLength)

	// Test Req Decode
	err := unbind.Decode(1, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, unbind.SequenceNumber[0] == 1)
	assert.True(t, unbind.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("unbind", unbind.Log()...)

	// Test ToResponse Rsp.Log
	unbindRsp := unbind.ToResponse(0)
	log.Info("unbindRsp", unbindRsp.Log()...)

	// Test Resp Encode
	dt = unbindRsp.Encode()
	log.Infof("data: %x", dt)
	rsp := unbindRsp.(*sgip.UnbindRsp)
	assert.True(t, uint32(len(dt)) == rsp.PacketLength)

	// Test Resp Decode
	err = unbindRsp.Decode(2, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, rsp.SequenceNumber[0] == 2)
	assert.True(t, rsp.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("unbindRsp", unbindRsp.Log()...)

}
