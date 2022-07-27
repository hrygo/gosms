package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIota(t *testing.T) {
	assert.True(t, OpFlowControl.String() == "flow_control")
}
