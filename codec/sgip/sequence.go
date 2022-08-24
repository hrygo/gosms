package sgip

import (
	"fmt"
	"sync"
	"time"
)

var Sequencer *sequence

type sequence struct {
	sync.Mutex
	node      uint32 // SP编号
	timestamp uint32 // 时间戳 mmddhhmmss 十进制数
	sequence  uint32 // 循序号
}

var seqInit sync.Once

func NewSequencer(node, worker uint32) *sequence {
	seqInit.Do(func() {
		Sequencer = &sequence{node: node, sequence: worker}
	})
	return Sequencer
}

func (s *sequence) NextVal() (rs []uint32) {
	s.Lock()
	defer s.Unlock()
	s.sequence = (s.sequence + 0x1f) & 0xffffffff
	now := time.Now()

	rs = make([]uint32, 3)
	rs[0] = s.node
	s.timestamp = uint32(now.Month() * 100000000)
	s.timestamp += uint32(now.Day() * 1000000)
	s.timestamp += uint32(now.Hour() * 10000)
	s.timestamp += uint32(now.Minute() * 100)
	s.timestamp += uint32(now.Second())
	rs[1] = s.timestamp
	rs[2] = s.sequence
	return
}

func (s *sequence) CurVal() (rs []uint32) {
	rs = make([]uint32, 3)
	rs[0] = s.node
	rs[1] = s.timestamp
	rs[2] = s.sequence
	return rs
}

func (s *sequence) String() string {
	return fmt.Sprintf("%010d%010d%08x", s.node, s.timestamp, s.sequence)
}
