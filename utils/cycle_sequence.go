package utils

import (
	"fmt"
	"sync"
)

// CycleSequence 生成可循环使用的序号
// 构成为: 0 | datacenter 2 bit | worker 3 bit| sequence 26 bit
// 最大支持32个节点，单节点2^26以内不会重复（67,108,864）
type CycleSequence struct {
	sync.Mutex       // 锁
	datacenter int32 // 数据中心机房id, 取值范围范围：0-4
	worker     int32 // 工作节点, 取值范围范围：0-8
	sequence   int32 // 序列号 26bit
}

const (
	sequenceMask    = int32(0x03ffffff)         // 最大值为26个1
	workerBits      = uint(3)                   // 机器id所占位数
	sequenceBits    = uint(28)                  // 序列所占的位数
	workerShift     = sequenceBits              // 机器id左移位数
	datacenterShift = sequenceBits + workerBits // 数据中心id左移位数
)

// NewCycleSequence d for datacenter-id, w for worker-id
func NewCycleSequence(d int32, w int32) *CycleSequence {
	return &CycleSequence{datacenter: d, worker: w}
}

func (s *CycleSequence) NextVal() int32 {
	s.Lock()
	defer s.Unlock()
	s.sequence = (s.sequence + 1) & sequenceMask
	r := (s.datacenter << datacenterShift) | (s.worker << workerShift) | (s.sequence)
	return r
}

func (s *CycleSequence) String() string {
	return fmt.Sprintf("%d:%d:%d", s.datacenter, s.worker, s.sequence)
}
