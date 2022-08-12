package sgip

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

func TestBind(t *testing.T) {
	// Test New Log
	bind := sgip.NewBind(ac, 1)
	log.Info("bind", bind.Log()...)

	// Test Req Encode
	dt := bind.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == bind.PacketLength)

	// Test Req Decode
	err := bind.Decode(1, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, bind.SequenceNumber[0] == 1)
	assert.True(t, bind.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("bind", bind.Log()...)

	// Test ToResponse Rsp.Log
	assert.True(t, bind.Check(ac) == 0)
	bindRsp := bind.ToResponse(uint32(bind.Check(ac)))
	log.Info("bindRsp", bindRsp.Log()...)

	// Test Resp Encode
	dt = bindRsp.Encode()
	log.Infof("data: %x", dt)
	rsp := bindRsp.(*sgip.BindRsp)
	assert.True(t, uint32(len(dt)) == rsp.PacketLength)

	// Test Resp Decode
	err = bindRsp.Decode(2, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, rsp.SequenceNumber[0] == 2)
	assert.True(t, rsp.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("bindRsp", bindRsp.Log()...)

}
