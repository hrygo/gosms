package smgp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

func TestLogin_Decode(t *testing.T) {
	lo := smgp.NewLogin(ac, uint32(codec.B32Seq.NextVal()))
	t.Logf("login   : %s", lo)
	assert.True(t, lo.ClientID() == ac.ClientId)

	lo.SetSecret(ac.SharedSecret)
	resp := lo.ToResponse(0).(*smgp.LoginRsp)
	t.Logf("resp    : %s", resp)
	assert.True(t, lo.ClientID() == ac.ClientId)

	dt1 := lo.Encode()
	dt2 := resp.Encode()
	assert.True(t, len(dt1) == smgp.LoginLen)
	assert.True(t, len(dt2) == smgp.LoginRespLen)

	err := lo.Decode(lo.SequenceId, dt1[12:])
	assert.True(t, err == nil)
	t.Logf("loginDec: %s, err: %v", lo, err)
	err = resp.Decode(resp.SequenceId, dt2[12:])
	assert.True(t, err == nil)
	t.Logf("respDec : %s, err: %v", resp, err)
}
