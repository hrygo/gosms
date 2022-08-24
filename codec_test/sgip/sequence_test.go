package sgip_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

func TestSeq(t *testing.T) {
	i := 100
	for i > 0 {
		i--
		assert.True(t, sgip.Sequencer.NextVal()[2]%0x1f == 0)
	}
}

func TestSequenceNumber_NextVal(t *testing.T) {
	rs := sgip.Sequencer.NextVal()
	t1 := fmt.Sprintf("%010d", rs[1])
	t2 := time.Now().Format("0102150405")
	assert.True(t, t1 == t2)
	assert.True(t, rs[2]%0x1f == 0)
}
