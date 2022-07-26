package codec

import (
	"encoding/binary"
	"errors"

	"github.com/hrygo/log"
)

// Protocol errors.

var ErrTotalLengthInvalid = errors.New("total_length in Packet data is invalid")
var ErrCommandIdInvalid = errors.New("command_id in Packet data is invalid")
var ErrCommandIdNotSupported = errors.New("command_id in Packet data is not supported")

type Logger interface {
	Log() []log.Field
}

type IHead interface {
	Logger
	Encode() []byte
	Decode([]byte) error
}

type Pdu interface {
	Logger
	Encode() []byte
	Decode(seq uint32, frame []byte) error
}

type RequestPdu interface {
	Pdu
	ToResponse(code uint32) Pdu
}

// OpError is the error type usually returned by functions in the packet
// package. It describes the operation and the error which the operation caused.
type OpError struct {
	// err is the error that occurred during the operation.
	// it is the origin error.
	err error

	// op is the operation which caused the error, such as
	// some "read" or "write" in packetWriter or packetReader.
	op string
}

func NewOpError(e error, op string) *OpError {
	return &OpError{
		err: e,
		op:  op,
	}
}

func (e *OpError) Error() string {
	if e.err == nil {
		return "<nil>"
	}
	return e.op + " error: " + e.err.Error()
}

func (e *OpError) Cause() error {
	return e.err
}

func (e *OpError) Op() string {
	return e.op
}

func UnpackHead(h []byte) (pkl, cmd, seq uint32) {
	if len(h) >= 12 {
		pkl = binary.BigEndian.Uint32(h[0:4])
		cmd = binary.BigEndian.Uint32(h[4:8])
		seq = binary.BigEndian.Uint32(h[8:12])
	}
	return
}
