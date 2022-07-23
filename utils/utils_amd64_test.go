package utils

import (
	"testing"
)

func TestIsBigEndian(t *testing.T) {
	b := IsBigEndian()
	if b {
		t.Errorf("The result of IsBigEndian is %v on Amd64 arch\n", b)
	}
}
