package smgp_test

import (
	"testing"

	"github.com/hrygo/gosms/bootstrap"
	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

var _ = bootstrap.BasePath

func TestActiveTest(t *testing.T) {
	at := smgp.NewActiveTest(uint32(codec.B32Seq.NextVal()))
	t.Logf("%T : %s", at, at)

	data := at.Encode()
	t.Logf("%T : %x", data, data)

	at2 := &smgp.ActiveTest{}
	_ = at2.Decode(uint32(codec.B32Seq.NextVal()), data)
	t.Logf("%T : %s", at2, at2)

	resp := at.ToResponse(0).(*smgp.ActiveTestRsp)
	t.Logf("%T : %s", resp, resp)

	data = resp.Encode()
	t.Logf("%T : %x", data, data)

	resp2 := &smgp.ActiveTest{}
	_ = resp2.Decode(uint32(codec.B32Seq.NextVal()), data)
	t.Logf("%T : %s", resp2, resp2)
}
