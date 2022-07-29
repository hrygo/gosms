package utils

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// TLV represents a Type-Length-Value object.
type TLV interface {
	Type() uint16
	Length() uint16
	Value() []byte
}

type object struct {
	typ uint16
	len uint16
	val []byte
}

// Type returns the object's type
func (o *object) Type() uint16 {
	return o.typ
}

// Length returns the object's type
func (o *object) Length() uint16 {
	return o.len
}

// Value returns the object's value
func (o *object) Value() []byte {
	return o.val
}

func (o *object) String() string {
	return fmt.Sprintf("{tag: %x, len: %d, val: %#x}", o.typ, o.len, o.val)
}

// Equal returns true if a pair of TLV objects are the same.
func Equal(tlv1, tlv2 TLV) bool {
	if tlv1 == nil {
		return tlv2 == nil
	} else if tlv2 == nil {
		return false
	} else if tlv1.Type() != tlv2.Type() {
		return false
	} else if tlv1.Length() != tlv2.Length() {
		return false
	} else if !bytes.Equal(tlv1.Value(), tlv2.Value()) {
		return false
	}
	return true
}

var (
	// ErrTLVRead is returned when there is an error reading a TLV object.
	ErrTLVRead = fmt.Errorf("TLV %s", "read error")
	// ErrTLVWrite is returned when  there is an error writing a TLV object.
	ErrTLVWrite = fmt.Errorf("TLV %s", "write error")
	// ErrTypeNotFound is returned when a request for a TLV type is made and none can be found.
	ErrTypeNotFound = fmt.Errorf("TLV %s", "type not found")
)

// New returns a TLV object from the args
func New(typ uint16, val []byte) TLV {
	tlv := new(object)
	tlv.typ = typ
	tlv.len = uint16(len(val))
	tlv.val = make([]byte, tlv.Length())
	copy(tlv.val, val)
	return tlv
}

// FromBytes returns a TLV object from bytes
func FromBytes(data []byte) (TLV, error) {
	objBuf := bytes.NewBuffer(data)
	return ReadObject(objBuf)
}

// ToBytes returns bytes from a TLV object
func ToBytes(tlv TLV) ([]byte, error) {
	data := make([]byte, 0)
	objBuf := bytes.NewBuffer(data)
	err := WriteObject(tlv, objBuf)
	return objBuf.Bytes(), err
}

// ReadObject returns a TLV object from io.Reader
func ReadObject(r io.Reader) (TLV, error) {
	tlv := new(object)

	var typ uint16
	var err error
	err = binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}
	tlv.typ = typ

	var length uint16
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	tlv.len = length

	tlv.val = make([]byte, tlv.Length())
	l, err := r.Read(tlv.val)
	if err != nil {
		return nil, err
	} else if uint16(l) != tlv.Length() {
		return tlv, ErrTLVRead
	}

	return tlv, nil
}

// WriteObject writes a TLV object to io.Writer
func WriteObject(tlv TLV, w io.Writer) error {
	var err error

	typ := tlv.Type()
	err = binary.Write(w, binary.BigEndian, typ)
	if err != nil {
		return err
	}

	length := tlv.Length()
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return err
	}

	n, err := w.Write(tlv.Value())
	if err != nil {
		return err
	} else if uint16(n) != tlv.Length() {
		return ErrTLVWrite
	}

	return nil
}

// TlvList is ad double-linked list containing TLV objects.
type TlvList struct {
	objects *list.List
}

// NewTlvList returns a new, empty TLVList.
func NewTlvList() *TlvList {
	tl := new(TlvList)
	tl.objects = list.New()
	return tl
}

// Length returns the number of objects int the TLVList.
func (tl *TlvList) Length() int32 {
	return int32(tl.objects.Len())
}

// Get checks the TLVList for any object matching the type, It returns the first one found.
// If the type could not be found, Get returns ErrTypeNotFound.
func (tl *TlvList) Get(typ uint16) (TLV, error) {
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		if e.Value.(*object).Type() == typ {
			return e.Value.(*object), nil
		}
	}
	return nil, ErrTypeNotFound
}

// GetAll checks the TLVList for all objects matching the type, returning a slice containing all matching objects.
// If no object has the requested type, an empty slice is returned.
func (tl *TlvList) GetAll(typ uint16) []TLV {
	ts := make([]TLV, 0)
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		if e.Value.(*object).Type() == typ {
			ts = append(ts, e.Value.(TLV))
		}
	}
	return ts
}

// Remove removes all objects with the requested type.
// It returns a count of the number of removed objects.
func (tl *TlvList) Remove(typ uint16) int {
	var totalRemoved int
	for {
		var removed int
		for e := tl.objects.Front(); e != nil; e = e.Next() {
			if e.Value.(*object).Type() == typ {
				tl.objects.Remove(e)
				removed++
				break
			}
		}
		if removed == 0 {
			break
		}
		totalRemoved += removed
	}
	return totalRemoved
}

// RemoveObject takes an TLV object as an argument, and removes all matching objects.
// It matches on not just type, but also the value contained in the object.
func (tl *TlvList) RemoveObject(obj TLV) int {
	var totalRemoved int
	for {
		var removed int
		for e := tl.objects.Front(); e != nil; e = e.Next() {
			if Equal(e.Value.(*object), obj) {
				tl.objects.Remove(e)
				removed++
				break
			}
		}

		if removed == 0 {
			break
		}
		totalRemoved += removed
	}
	return totalRemoved
}

// Add pushes a new TLV object onto the TLVList. It builds the object from its args
func (tl *TlvList) Add(typ uint16, value []byte) {
	obj := New(typ, value)
	tl.objects.PushBack(obj)
}

// AddObject adds a TLV object onto the TLVList
func (tl *TlvList) AddObject(obj TLV) {
	tl.objects.PushBack(obj)
}

// Write writes out the TLVList to an io.Writer.
func (tl *TlvList) Write(w io.Writer) error {
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		err := WriteObject(e.Value.(TLV), w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tl *TlvList) String() string {
	if tl == nil {
		return "[]"
	}
	var sb strings.Builder
	sb.Grow(int(8 * tl.Length()))
	sb.WriteString("[")
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		o := e.Value.(*object)
		sb.WriteString(o.String())
		sb.WriteString(",")
	}
	sb.WriteString("]")
	return sb.String()
}

// Read takes an io.Reader and builds a TLVList from that.
func Read(r io.Reader) (*TlvList, error) {
	tl := NewTlvList()
	var err error
	for {
		var tlv TLV
		if tlv, err = ReadObject(r); err != nil {
			break
		}
		tl.objects.PushBack(tlv)
	}

	if err == io.EOF {
		err = nil
	}
	return tl, err
}
