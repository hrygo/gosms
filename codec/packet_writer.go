package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type packetWriter struct {
	wb  *bytes.Buffer
	err *OpError
}

func NewPacketWriter(initSize uint32) *packetWriter {
	buf := make([]byte, 0, initSize)
	return &packetWriter{
		wb: bytes.NewBuffer(buf),
	}
}

// Bytes returns a slice of the contents of the inner buffer;
// If the caller changes the contents of the
// returned slice, the contents of the buffer will change provided there
// are no intervening method calls on the Buffer.
func (w *packetWriter) Bytes() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	l := w.wb.Len()
	return (w.wb.Bytes())[:l], nil
}

// WriteByte appends the byte of b to the inner buffer, growing the buffer as
// needed.
func (w *packetWriter) WriteByte(b byte) {
	if w.err != nil {
		return
	}

	err := w.wb.WriteByte(b)
	if err != nil {
		w.err = NewOpError(err, fmt.Sprintf("packetWriter.WriteByte writes: %x", b))
		return
	}
}

func (w *packetWriter) WriteBytes(bs []byte) {
	if w.err != nil {
		return
	}

	l1 := len(bs)
	l2 := l1
	if l2 > 10 {
		l2 = 10
	}

	n, err := w.wb.Write(bs)
	if err != nil {
		w.err = NewOpError(err, fmt.Sprintf("packetWriter.WriteByte writes: %x", bs))
		return
	}

	if l1 != n {
		w.err = NewOpError(ErrMethodParamsInvalid,
			fmt.Sprintf("packetWriter.WriteBytes writes: %x", bs[0:l2]))
		return
	}
}

// WriteFixedSizeString writes a string to buffer, if the length of s is less than size,
// Pad binary zero to the right.
func (w *packetWriter) WriteFixedSizeString(s string, size int) {
	if w.err != nil {
		return
	}

	l1 := len(s)
	l2 := l1
	if l2 > 10 {
		l2 = 10
	}

	if l1 > size {
		w.err = NewOpError(ErrMethodParamsInvalid,
			fmt.Sprintf("packetWriter.WriteFixedSizeString writes: %s", s[0:l2]))
		return
	}

	w.WriteString(strings.Join([]string{s, string(make([]byte, size-l1))}, ""))
}

// WriteString appends the contents of s to the inner buffer, growing the buffer as
// needed.
func (w *packetWriter) WriteString(s string) {
	if w.err != nil {
		return
	}

	l1 := len(s)
	l2 := l1
	if l2 > 10 {
		l2 = 10
	}

	n, err := w.wb.WriteString(s)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("packetWriter.WriteString writes: %s...", s[0:l2]))
		return
	}

	if n != l1 {
		w.err = NewOpError(fmt.Errorf("WriteString writes %d bytes, not equal to %d we expected", n, l1),
			fmt.Sprintf("packetWriter.WriteString writes: %s...", s[0:l2]))
		return
	}
}

// WriteInt appends the content of data to the inner buffer in order, growing the buffer as
// needed.
func (w *packetWriter) WriteInt(order binary.ByteOrder, data interface{}) {
	if w.err != nil {
		return
	}

	err := binary.Write(w.wb, order, data)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("packetWriter.WriteInt writes: %#v", data))
		return
	}
}
