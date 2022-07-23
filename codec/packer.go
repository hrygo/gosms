package codec

import (
	"encoding/binary"
	"errors"
)

// Common errors.

var ErrMethodParamsInvalid = errors.New("params passed to method is invalid")

// Protocol errors.

var ErrTotalLengthInvalid = errors.New("total_length in Packet data is invalid")
var ErrCommandIdInvalid = errors.New("command_id in Packet data is invalid")
var ErrCommandIdNotSupported = errors.New("command_id in Packet data is not supported")

type Packer interface {
	Pack(seqId uint32) ([]byte, error)
	Unpack(data []byte) error
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

func UnpackHead(h []byte) (pkl, cmd uint32) {
	if len(h) < 8 {
		return 0, 0
	} else {
		pkl = binary.BigEndian.Uint32(h[0:4])
		cmd = binary.BigEndian.Uint32(h[4:8])
		return
	}
}
