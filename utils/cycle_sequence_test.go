package utils_test

import (
	"testing"

	"github.com/hrygo/gosmsn/utils"
)

var seq = utils.NewCycleSequence(1, 1)

func BenchmarkCycleSequence_NextVal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seq.NextVal()
	}
}
