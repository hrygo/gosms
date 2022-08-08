package utils_test

import (
	"fmt"
	"testing"

	"github.com/hrygo/gosms/utils"
)

func Test_BcdToString(t *testing.T) {
	bcd := []byte{0x01, 0x23, 0x45, 0x67, 0x8a}
	str := utils.BcdToString(bcd)

	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
	t.Logf("str: %s", str)
}

func Test_StoBcd(t *testing.T) {
	str := "012345678a"
	bcd := utils.StoBcd(str)
	t.Logf("str: %s", str)
	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
}

var bcdSeq = utils.NewBcdSequence("000001")

func BenchmarkBcdSequence_NextSeq(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcdSeq.NextVal()
	}
}

func BenchmarkBcdSequence_BcdToString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.BcdToString(bcdSeq.NextVal())
	}
}

func BenchmarkBcdSequence_FmtString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%x", bcdSeq.NextVal())
	}
}

func BenchmarkBcdSequence_BcdToStringParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			utils.BcdToString(bcdSeq.NextVal())
		}
	})
}

func BenchmarkBcdSequence_FmtStringParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fmt.Sprintf("%x", bcdSeq.NextVal())
		}
	})
}
