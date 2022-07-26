package smgp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/smgp"
)

func TestNewSubmit(t *testing.T) {
	subs := smgp.NewSubmit(ac, []string{"17600001111", "17700001111"}, Poem, uint32(codec.B32Seq.NextVal()))
	assert.True(t, len(subs) == 4)

	for i, sub := range subs {
		sub := sub.(*smgp.Submit)
		t.Logf("%+v", sub)
		if i < 3 {
			assert.True(t, int(sub.PacketLength) == 328)
			assert.True(t, int(sub.MsgLength()) == 140)
			assert.True(t, int(sub.MsgLength()) == len(sub.MsgContent()))
		} else {
			assert.True(t, int(sub.MsgLength()) <= 140)
			assert.True(t, int(sub.PacketLength) > 147)
		}
	}
}

func TestSubmit_Encode(t *testing.T) {
	encode(t, []string{"17600001111", "17600002222"}, Poem, 4)
	encode(t, []string{"17600001111"}, "hello world 世界，你好！", 1)
}

func encode(t *testing.T, phones []string, txt string, l int) {
	subs := smgp.NewSubmit(ac, phones, txt, uint32(codec.B32Seq.NextVal()))
	assert.True(t, len(subs) == l)

	for _, sub := range subs {
		sub := sub.(*smgp.Submit)
		t.Logf("%+v", sub)
		dt := sub.Encode()
		assert.True(t, int(sub.PacketLength) == len(dt))
		t.Logf("%v: %x", int(sub.PacketLength) == len(dt), dt)
		resp := sub.ToResponse(0).(*smgp.SubmitRsp)
		t.Logf("%s", resp)
		dt = resp.Encode()
		t.Logf("%v: %x", int(resp.PacketLength) == len(dt), dt)
	}
}

func TestSubmit_Decode(t *testing.T) {
	decode(t, []string{"17600001111", "17600002222"}, Poem, 4)
	decode(t, []string{"17600001111"}, "hello world 世界，你好！", 1)
}

func decode(t *testing.T, phones []string, txt string, l int) {
	subs := smgp.NewSubmit(ac, phones, txt, uint32(codec.B32Seq.NextVal()))
	assert.True(t, len(subs) == l)

	for _, sub := range subs {
		sub := sub.(*smgp.Submit)
		dt := sub.Encode()
		assert.True(t, int(sub.PacketLength) == len(dt))
		t.Logf("%s", sub)
		t.Logf("%v: %x", int(sub.PacketLength) == len(dt), dt)

		subDec := &smgp.Submit{}
		err := subDec.Decode(sub.SequenceId, dt[12:])
		if err != nil {
			t.Fail()
			continue
		}
		dt2 := subDec.Encode()
		t.Logf("%s", subDec)
		t.Logf("%v: %x", int(subDec.PacketLength) == len(dt2), dt2)
		content := sub.MsgContent()
		if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
			content = content[6:]
		}
		ctx1, _ := smgp.GbDecoder.Bytes(content)
		ctx2 := subDec.MsgContent()
		assert.True(t, bytes.Equal(ctx1, ctx2))

		resp := subDec.ToResponse(0).(*smgp.SubmitRsp)
		t.Logf("%s", resp)
		dt = resp.Encode()
		assert.True(t, int(resp.PacketLength) == len(dt))
		t.Logf("%v: %x", int(resp.PacketLength) == len(dt), dt)

		respDec := &smgp.SubmitRsp{}
		err = respDec.Decode(sub.SequenceId, dt[12:])
		if err != nil {
			t.Fail()
			continue
		}
		t.Logf("%s", respDec)
		assert.True(t, int(respDec.PacketLength) == len(dt))
		t.Logf("%v: %x", int(respDec.PacketLength) == len(dt), dt)
		assert.True(t, 0 == bytes.Compare(respDec.MsgId(), resp.MsgId()))
	}
}

func TestGbk(t *testing.T) {
	gb, _ := smgp.GbEncoder.String(Poem)
	gbDec, _ := smgp.GbDecoder.String(gb)
	t.Logf("Origin: %s", Poem)
	t.Logf("GbStr : %s", gbDec)
	t.Logf("Origin Hex: %x", Poem)
	t.Logf("GbStr  Hex: %x", gb)
}

const Poem = "将进酒\n" +
	"君不见黄河之水天上来，奔流到海不复回。\n" +
	"君不见高堂明镜悲白发，朝如青丝暮成雪。\n" +
	"人生得意须尽欢，莫使金樽空对月。\n" +
	"天生我材必有用，千金散尽还复来。\n" +
	"烹羊宰牛且为乐，会须一饮三百杯。\n" +
	"岑夫子，丹丘生，将进酒，杯莫停。\n" +
	"与君歌一曲，请君为我倾耳听。\n" +
	"钟鼓馔玉不足贵，但愿长醉不愿醒。\n" +
	"古来圣贤皆寂寞，惟有饮者留其名。\n" +
	"陈王昔时宴平乐，斗酒十千恣欢谑。\n" +
	"主人何为言少钱，径须沽取对君酌。\n" +
	"五花马、千金裘，呼儿将出换美酒，与尔同销万古愁。"
