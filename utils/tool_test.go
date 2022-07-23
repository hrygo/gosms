package utils

import (
  "testing"

  "github.com/stretchr/testify/assert"
  "golang.org/x/text/encoding/simplifiedchinese"
  "golang.org/x/text/encoding/unicode"
)

func TestUcs2Encode(t *testing.T) {
  var s = "【中原银行】您尾号0045的0054于10月23日01:57在电费缴费100,000,000元，可用余额为23,456.00元。客服电话：95186。"
  t.Logf("%s", s)
  t.Logf("%x", s)

  s, _ = unicode.UTF8.NewEncoder().String(s)
  t.Logf("%s", s)
  t.Logf("%x", s)

  t.Logf("%x", Ucs2Encode(s))

  s, _ = simplifiedchinese.GB18030.NewEncoder().String(s)
  t.Logf("%x", s)
  s, _ = simplifiedchinese.GB18030.NewDecoder().String(s)
  t.Logf("%s", s)
  t.Logf("%x", s)
}

func TestDiceCheck(t *testing.T) {
  assert.False(t, DiceCheck(1))
}
