package smgp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

func TestReport(t *testing.T) {
	id := codec.BcdSeq.NextVal()
	rpt := smgp.NewReport(id)
	t.Logf("rpt: %s", rpt)
	data := rpt.Encode()
	assert.True(t, len(data) == smgp.RptLen)
	t.Logf("value: %x", data)

	rpt2 := &smgp.Report{}
	err := rpt2.Decode(data)
	assert.True(t, err == nil)
	t.Logf("rpt2: %s", rpt2)
}
