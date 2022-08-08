package snowflake32_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/utils/snowflake32"
)

var seq32 = snowflake32.NewSnowflake(3, 7)

func TestNewSnowflake(t *testing.T) {
	assert.True(t, seq32 != nil)
}

func TestSnowflake_NextVal(t *testing.T) {
	a, b := seq32.NextVal(), seq32.NextVal()
	assert.True(t, a < b)
	assert.True(t, b-a == 1)
}

func TestSnowflake_String(t *testing.T) {
	sf := seq32
	seq := sf.NextVal()
	assert.True(t, seq&0x1ff == sf.Sequence())

	for (seq & 0x1ff) < 0x00f {
		seq = sf.NextVal()
		t.Log(sf)
	}

	sfn := snowflake32.Parse(seq)
	t.Log(sfn)

	assert.True(t, sf.Seconds() == sfn.Seconds())
	assert.True(t, sf.Datacenter() == sfn.Datacenter())
	assert.True(t, sf.Worker() == sfn.Worker())
	assert.True(t, sf.Sequence() == sfn.Sequence())
}
