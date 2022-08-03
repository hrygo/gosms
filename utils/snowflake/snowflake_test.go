package snowflake_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils/snowflake"
)

var _ = bootstrap.BasePath

func TestNewSnowflake(t *testing.T) {
	assert.True(t, codec.B64Seq != nil)
}

func TestSnowflake_NextVal(t *testing.T) {
	a, b := codec.B64Seq.NextVal(), codec.B64Seq.NextVal()
	assert.True(t, a < b)
	assert.True(t, b-a == 1)
}

func TestSnowflake_String(t *testing.T) {
	sf := codec.B64Seq.(*snowflake.Snowflake)
	seq := sf.NextVal()
	assert.True(t, seq&0xfff == sf.Sequence())

	for (seq & 0xfff) < 0x00f {
		seq = sf.NextVal()
		t.Log(sf)
	}

	sfn := snowflake.Parse(seq)
	t.Log(sfn)

	assert.True(t, sf.Timestamp() == sfn.Timestamp())
	assert.True(t, sf.DatacenterId() == sfn.DatacenterId())
	assert.True(t, sf.WorkerId() == sfn.WorkerId())
	assert.True(t, sf.Sequence() == sfn.Sequence())
}
