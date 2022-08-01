package cmpp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
)

var _ = bootstrap.BasePath
var cli = auth.Cache.FindByCid("cmpp", "123456")

func TestCmppConnect_Encode(t *testing.T) {
	connect := NewConnect(cli, uint32(codec.B32Seq.NextVal()))
	t.Logf("%v", connect)

	frame := connect.Encode()
	t.Logf("Connect: %x", frame)
	assert.Equal(t, uint32(0), uint32(connect.Check(cli)))

	connect.SetSecret(cli.SharedSecret)
	resp := connect.ToResponse(0).(*ConnectResp)
	t.Logf("Connect: %v", resp)
	t.Logf("ConnectResp: %x", resp.Encode())
}
