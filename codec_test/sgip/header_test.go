package sgip

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

var node uint32 = 3037196688
var seq = sgip.SequenceNumber{Node: node}

func TestSequenceNumber_CurVal(t *testing.T) {
	t.Logf("%x", seq.CurVal())
	assert.True(t, seq.CurVal()[0] == node)
	assert.True(t, seq.CurVal()[1] == 0)
	assert.True(t, seq.CurVal()[2] == 0)
}

func TestSequenceNumber_NextVal(t *testing.T) {
	rs := seq.NextVal()
	assert.True(t, rs[0] == node)
	t1 := fmt.Sprintf("%010d", rs[1])
	t2 := time.Now().Format("0102150405")
	assert.True(t, t1 == t2)
	assert.True(t, rs[2] == 1)
}

func TestSequenceNumber_String(t *testing.T) {
	seq.NextVal()
	t.Logf("%s", &seq)
}
