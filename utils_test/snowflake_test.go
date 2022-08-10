package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/utils/snowflake"
)

var seq64 = snowflake.NewSnowflake(5, 127)

func TestNewSnowflake(t *testing.T) {
	assert.True(t, seq64 != nil)
}

func TestSnowflake_NextVal(t *testing.T) {
	a, b := seq64.NextVal(), seq64.NextVal()
	assert.True(t, a < b)
	assert.True(t, b-a == 1)
}

func TestSnowflake_String(t *testing.T) {
	seq := seq64.NextVal()
	assert.True(t, seq&0xfff == seq64.Sequence())

	for (seq & 0xfff) < 0x00f {
		seq = seq64.NextVal()
		t.Log(seq64)
	}

	sfn := snowflake.Parse(seq)
	t.Log(sfn)

	assert.True(t, seq64.Timestamp() == sfn.Timestamp())
	assert.True(t, seq64.DatacenterId() == sfn.DatacenterId())
	assert.True(t, seq64.WorkerId() == sfn.WorkerId())
	assert.True(t, seq64.Sequence() == sfn.Sequence())
}
