package smgp_test

import (
	"testing"

	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/smgp"
)

var _ = bootstrap.BasePath

func TestActiveTest(t *testing.T) {
	at := smgp.NewActiveTest()
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
