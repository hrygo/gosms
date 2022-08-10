package cmpp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/cmpp"
)

func TestCmppConnect_Encode(t *testing.T) {
	connect := cmpp.NewConnect(ac, uint32(codec.B32Seq.NextVal()))
	t.Logf("%v", connect)

	frame := connect.Encode()
	t.Logf("Connect: %x", frame)
	assert.Equal(t, uint32(0), uint32(connect.Check(ac)))

	connect.SetSecret(ac.SharedSecret)
	resp := connect.ToResponse(0).(*cmpp.ConnectResp)
	t.Logf("Connect: %v", resp)
	t.Logf("ConnectResp: %x", resp.Encode())
}
