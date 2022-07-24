package codec

// Sequence32 32位序号生成器
type Sequence32 interface {
	NextVal() int32
}

// Sequence64 64位序号生成器
type Sequence64 interface {
	NextVal() int64
}

// SequenceBCD BCD码序号生成器
type SequenceBCD interface {
	NextVal() []byte
}

var B32Seq Sequence32
var B64Seq Sequence64
var BcdSeq SequenceBCD
