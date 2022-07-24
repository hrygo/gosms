package snowflake32

import (
	"fmt"
	"sync"
	"time"
)

// Snowflake 24小时内不会重复的雪花序号生成器
// 构成为: 0 | seconds 17 bit | datacenter 2 bit | worker 3 bit| sequence 9 bit
// 最大支持32个节点，单节点TPS不超过512，超过则会阻塞程序到下一秒再返回序号，仅能用于特殊场景
// seconds占用17bits是因为一天86400秒占用17bits
type Snowflake struct {
	sync.Mutex       // 锁
	seconds    int32 // 时间戳 ，截止到午夜0点的秒数
	datacenter int32 // 数据中心机房id, 取值范围范围：0-4
	worker     int32 // 工作节点, 取值范围范围：0-8
	sequence   int32 // 序列号
}

const (
	sequenceMask    = int32(0x01ff)                              // 最大值为9个1
	datacenterBits  = uint(2)                                    // 数据中心id所占位数
	workerBits      = uint(3)                                    // 机器id所占位数
	sequenceBits    = uint(9)                                    // 序列所占的位数
	workerShift     = sequenceBits                               // 机器id左移位数
	datacenterShift = sequenceBits + workerBits                  // 数据中心id左移位数
	timestampShift  = sequenceBits + workerBits + datacenterBits // 时间戳左移位数
)

// NewSnowflake d for datacenter-id, w for worker-id
func NewSnowflake(d int32, w int32) *Snowflake {
	return &Snowflake{datacenter: d, worker: w}
}

func (s *Snowflake) NextVal() int32 {
	s.Lock()
	defer s.Unlock()
	now := passedSeconds() // 获得当前秒
	if s.seconds == now {
		// 当同一时间戳（精度：秒）下次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出9bit长度，则需要等待下一秒
			// 下一秒将使用sequence:0
			for now <= s.seconds {
				time.Sleep(time.Microsecond)
				now = passedSeconds()
			}
		}
	} else {
		// 不同时间戳（精度：秒）下直接使用序列号：0
		s.sequence = 0
	}
	s.seconds = now
	r := (s.seconds << timestampShift) | (s.datacenter << datacenterShift) | (s.worker << workerShift) | (s.sequence)
	return r
}

func (s *Snowflake) String() string {
	return fmt.Sprintf("%d:%d:%d:%d", s.seconds, s.datacenter, s.worker, s.sequence)
}

func passedSeconds() int32 {
	t := time.Now()
	return int32(t.Hour()*3600 + t.Minute()*60 + t.Second())
}
