package sgip

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec/sgip"
)

func TestReport(t *testing.T) {
	// Test New Log
	pdu := sgip.NewReport("18600001111", sgip.Sequencer.NextVal(), sgip.Status(0), 2)
	log.Info("report", pdu.Log()...)
	report := pdu.(*sgip.Report)

	// Test Req Encode
	dt := report.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == report.PacketLength)

	// Test Req Decode
	err := report.Decode(1, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, report.SequenceNumber[0] == 1)
	assert.True(t, report.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("report", report.Log()...)

	// Test ToResponse Rsp.Log
	pdu2 := report.ToResponse(0)
	log.Info("reportRsp", pdu2.Log()...)
	reportRsp := pdu2.(*sgip.ReportRsp)

	// Test Resp Encode
	dt = reportRsp.Encode()
	log.Infof("data: %x", dt)
	assert.True(t, uint32(len(dt)) == reportRsp.PacketLength)

	// Test Resp Decode
	err = reportRsp.Decode(2, dt[12:])
	if err != nil {
		t.Error(err)
	}
	assert.True(t, err == nil)
	assert.True(t, reportRsp.SequenceNumber[0] == 2)
	assert.True(t, reportRsp.SequenceNumber[1] == binary.BigEndian.Uint32(dt[12:]))
	log.Info("reportRsp", reportRsp.Log()...)

}
