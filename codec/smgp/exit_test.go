package smgp_test

import (
	"testing"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/smgp"
)

func TestExit(t *testing.T) {
	exit := smgp.NewExit(uint32(codec.B32Seq.NextVal()))
	t.Logf("%T : %s", exit, exit)

	data := exit.Encode()
	t.Logf("%T : %x", data, data)

	e2 := &smgp.Exit{}
	_ = e2.Decode(uint32(codec.B32Seq.NextVal()), data)
	t.Logf("%T : %s", e2, e2)

	resp := exit.ToResponse(0).(*smgp.ExitRsp)
	t.Logf("%T : %s", resp, resp)

	data = resp.Encode()
	t.Logf("%T : %x", data, data)

	resp2 := &smgp.ExitRsp{}
	_ = resp2.Decode(uint32(codec.B32Seq.NextVal()), data)
	t.Logf("%T : %s", resp2, resp2)
}
