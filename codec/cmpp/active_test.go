package cmpp_test

import (
	"testing"

	"github.com/hrygo/gosmsn/codec/cmpp"
)

func TestActiveTestReqPktPack(t *testing.T) {
	p := &cmpp.ActiveTestReqPkt{}

	data, err := p.Pack(seqId)
	if err != nil {
		t.Fatal("ActiveTestReqPkt pack error:", err)
	}

	if p.SeqId != seqId {
		t.Fatalf("After pack, seqId is %d, not equal to expected: %d\n", p.SeqId, seqId)
	}

	// data after pack expected:
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17,
	}

	l1 := len(data)
	l2 := len(dataExpected)
	if l1 != l2 {
		t.Fatalf("After pack, data length is %d, not equal to length expected: %d\n", l1, l2)
	}

	for i := 0; i < l1; i++ {
		if data[i] != dataExpected[i] {
			t.Fatalf("After pack, data[%d] is %x, not equal to dataExpected[%d]: %x\n", i, data[i], i, dataExpected[i])
		}
	}
}

func TestActiveTestReqUnpack(t *testing.T) {
	// cmpp active test request packet data:
	data := []byte{
		0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17,
	}

	p := &cmpp.ActiveTestReqPkt{}
	err := p.Unpack(data[8:])
	if err != nil {
		t.Fatal("ActiveTestReqPkt unpack error:", err)
	}

	if p.SeqId != seqId {
		t.Fatalf("After unpack, seqId in packet is %x, not equal to the expected value: %x\n", p.SeqId, seqId)
	}
}

func TestActiveTestRspPktPack(t *testing.T) {
	p := &cmpp.ActiveTestRspPkt{}

	data, err := p.Pack(seqId)
	if err != nil {
		t.Fatal("ActiveTestRspPkt pack error:", err)
	}

	if p.SeqId != seqId {
		t.Fatalf("After pack, seqId is %d, not equal to expected: %d\n", p.SeqId, seqId)
	}

	// data after pack expected:
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x0d, 0x80, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17, 0x00,
	}

	l1 := len(data)
	l2 := len(dataExpected)
	if l1 != l2 {
		t.Fatalf("After pack, data length is %d, not equal to length expected: %d\n", l1, l2)
	}

	for i := 0; i < l1; i++ {
		if data[i] != dataExpected[i] {
			t.Fatalf("After pack, data[%d] is %x, not equal to dataExpected[%d]: %x\n", i, data[i], i, dataExpected[i])
		}
	}
}

func TestActiveTestRspUnpack(t *testing.T) {
	// cmpp active test response packet data:
	data := []byte{
		0x00, 0x00, 0x00, 0x0d, 0x80, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17, 0x00,
	}

	p := &cmpp.ActiveTestRspPkt{}
	err := p.Unpack(data[8:])
	if err != nil {
		t.Fatal("ActiveTestRspPkt unpack error:", err)
	}

	if p.SeqId != seqId {
		t.Fatalf("After unpack, seqId in packet is %x, not equal to the expected value: %x\n", p.SeqId, seqId)
	}
}
