package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var ErrInvalidUtf8Rune = errors.New("Not Invalid Utf8 runes")

func IsBigEndian() bool {
	var i uint16 = 0x1234
	var p = (*[2]byte)(unsafe.Pointer(&i))
	if (*p)[0] == 0x12 {
		return true
	}
	return false
}

func Now() (string, uint32) {
	s := time.Now().Format("0102150405")
	i, _ := strconv.Atoi(s)
	return s, uint32(i)
}

// TimeStamp2Str converts a timestamp(MMDDHHMMSS) int to a string(10 bytes).
func TimeStamp2Str(t uint32) string {
	return fmt.Sprintf("%010d", t)
}

func Utf8ToUcs2(in string) ([]byte, error) {
	if !utf8.ValidString(in) {
		return nil, ErrInvalidUtf8Rune
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()) // UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Ucs2ToUtf8(in []byte) (string, error) {
	r := bytes.NewReader(in)
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()) // UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Utf8ToGB18030(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", ErrInvalidUtf8Rune
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewEncoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GB18030ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewDecoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func OctetString(s string, fixedLength int) string {
	length := len(s)
	if length == fixedLength {
		return s
	}

	if length > fixedLength {
		return s[length-fixedLength:]
	}

	return strings.Join([]string{s, string(make([]byte, fixedLength-length))}, "")
}

func TrimOctetString(in []byte) []byte {
	i := bytes.IndexByte(in, 0)
	if i == -1 {
		return in
	} else {
		return in[:i]
	}
}

func TrimStr(in []byte) string {
	return string(TrimOctetString(in))
}

// ToTPUDHISlices 拆分为长短信切片
// 纯ASCII内容的拆分 pkgLen = 160
// 含中文内容的拆分   pkgLen = 140
func ToTPUDHISlices(content []byte, pkgLen int) (rt [][]byte) {
	if len(content) < pkgLen {
		return [][]byte{content}
	}

	headLen := 6
	bodyLen := pkgLen - headLen
	parts := len(content) / bodyLen
	tailLen := len(content) % bodyLen
	if tailLen != 0 {
		parts++
	}
	// 分片消息组的标识，用于收集组装消息
	groupId := byte(time.Now().UnixNano() & 0xff)
	var part []byte
	for i := 0; i < parts; i++ {
		if i != parts-1 {
			part = make([]byte, pkgLen)
		} else {
			// 最后一片
			part = make([]byte, 6+tailLen)
		}
		part[0], part[1], part[2] = 0x05, 0x00, 0x03
		part[3] = groupId
		part[4], part[5] = byte(parts), byte(i+1)
		if i != parts-1 {
			copy(part[6:pkgLen], content[bodyLen*i:bodyLen*(i+1)])
		} else {
			copy(part[6:], content[0:tailLen])
		}
		rt = append(rt, part)
	}
	return rt
}

func RandNum(min, max int) int {
	return rand.Intn(max-min) + min
}

// DiceCheck 投概率骰子，得到结果比给定数字大则返回true，否则返回false
func DiceCheck(prob float64) bool {
	return float64(rand.Intn(10000))/10000.0 > prob
}
