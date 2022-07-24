package utils

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// BcdSequence 电信MsgId序号生成器，支持每分钟产生100万个不重复序号，超出后序号会重复
// BCD 4bit编码，用4bit表示0-9的数字
type BcdSequence struct {
	sync.Mutex        // 锁
	worker     []byte // SMGW代码：3 字节（ BCD 码），6位十进制数字的字符串
	timestamp  string // 时间：4 字节（ BCD 码），格式为 MMDDHHMM（月日时分）
	sequence   int32  // 序列号：3 字节（ BCD 码），取值范围为 000000 999999 ，从 0 开始，顺序累加，步长为 1 循环使用。
}

const telSeqMax = 1000000

func NewBcdSequence(w string) *BcdSequence {
	// check
	for _, s := range w {
		if byte(s) > '9' || byte(s) < '0' {
			w = "000000"
			break
		}
	}
	w = "000000" + w
	w = w[len(w)-6:]

	ret := &BcdSequence{}
	ret.worker = StoBcd(w)
	return ret
}

func (tf *BcdSequence) NextVal() []byte {
	tf.Lock()
	defer tf.Unlock()
	mi := time.Now().Format("01021504")
	if tf.timestamp == mi {
		// 超过每分钟telSeqMax后序号会重复
		tf.sequence = (tf.sequence + 1) % telSeqMax
	} else {
		tf.sequence = 0
	}
	tf.timestamp = mi
	seq := make([]byte, 10)
	copy(seq[0:3], tf.worker)
	copy(seq[3:7], StoBcd(tf.timestamp))
	copy(seq[7:10], StoBcd(IntToFixStr(int64(tf.sequence), 6)))
	return seq
}

func IntToFixStr(i int64, l int) string {
	si := strconv.FormatInt(i, 10)
	if len(si) == l {
		return si
	} else {
		var sb strings.Builder
		sb.Grow(l)
		for i := 0; i < l; i++ {
			sb.WriteByte('0')
		}
		sb.WriteString(si)
		si = sb.String()
		return si[len(si)-l:]
	}
}

func StoBcd(w string) []byte {
	var wb []byte
	var h, l byte
	for i, c := range []byte(w) {
		// index 为偶数的作为高4bit
		if i&0x1 == 0 {
			h = c - '0'
			if h > 9 {
				h = 9
			}
		} else {
			// index 为奇数的作为低4bit
			l = c - '0'
			if l > 9 {
				l = 9
			}
		}
		// 每两个字符构成一个字节
		if i&0x1 == 1 {
			wb = append(wb, h<<4|l)
			h, l = 0, 0
		}
	}
	if h != 0 {
		wb = append(wb, h<<4)
	}
	return wb
}

func BcdToString(bcd []byte) string {
	var sb strings.Builder
	sb.Grow(2 * len(bcd))
	for _, b := range bcd {
		c := b >> 4
		if c > 9 {
			c = 9
		}
		sb.WriteByte(c + '0')
		c = b & 0x0f
		if c > 9 {
			c = 9
		}
		sb.WriteByte(c + '0')
	}
	return sb.String()
}
