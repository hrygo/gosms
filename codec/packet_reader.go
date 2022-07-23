package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const maxCStringSize = 160

type packetReader struct {
	rb   *bytes.Buffer
	err  *OpError
	buff [maxCStringSize]byte
}

func NewPacketReader(data []byte) *packetReader {
	return &packetReader{
		rb: bytes.NewBuffer(data),
	}
}

// ReadByte reads and returns the next byte from the inner buffer.
// If no byte is available, it returns an OpError.
func (r *packetReader) ReadByte() byte {
	if r.err != nil {
		return 0
	}

	b, err := r.rb.ReadByte()
	if err != nil {
		r.err = NewOpError(err,
			"packetReader.ReadByte")
		return 0
	}
	return b
}

// ReadInt reads structured binary data from r into data.
// Data must be a pointer to a fixed-size value or a slice
// of fixed-size values.
// Bytes read from r are decoded using the specified byte order
// and written to successive fields of the data.
func (r *packetReader) ReadInt(order binary.ByteOrder, data interface{}) {
	if r.err != nil {
		return
	}

	err := binary.Read(r.rb, order, data)
	if err != nil {
		r.err = NewOpError(err,
			"packetReader.ReadInt")
		return
	}
}

// ReadBytes reads the next len(s) bytes from the inner buffer to s.
// If the buffer has no data to return, an OpError would be stored in r.err.
func (r *packetReader) ReadBytes(s []byte) {
	if r.err != nil {
		return
	}

	n, err := r.rb.Read(s)
	if err != nil {
		r.err = NewOpError(err,
			"packetReader.ReadBytes")
		return
	}

	if n != len(s) {
		r.err = NewOpError(fmt.Errorf("ReadBytes reads %d bytes, not equal to %d we expected", n, len(s)),
			"packetWriter.ReadBytes")
		return
	}
}

// ReadCString read bytes from packerReader's inner buffer,
// it would trim the tail-zero byte and the bytes after that.
// before next ReadCString() you mast copy the result value(convert to target string).
func (r *packetReader) ReadCString(length int) []byte {
	if r.err != nil {
		return nil
	}

	var tmp = r.buff[:length]
	n, err := r.rb.Read(tmp)
	if err != nil {
		r.err = NewOpError(err,
			"packetReader.ReadCString")
		return nil
	}

	if n != length {
		r.err = NewOpError(fmt.Errorf("ReadCString reads %d bytes, not equal to %d we expected", n, length),
			"packetWriter.ReadCString")
		return nil
	}

	i := bytes.IndexByte(tmp, 0)
	if i == -1 {
		return tmp
	} else {
		return tmp[:i]
	}
}

// Error return the inner err.
func (r *packetReader) Error() error {
	if r.err != nil {
		return r.err
	}
	return nil
}
