package cmpp

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/cmpp"
	"github.com/hrygo/gosms/utils"
)

func TestEncode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"testcase1",
			args{"1"},
			[]byte{0x00, 0x31},
		},
		{
			"testcase2",
			args{"hello world"},
			[]byte{0x00, 0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00, 0x20, 0x00, 0x77, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6c, 0x00, 0x64},
		},
		{"testcase3",
			args{"Great 中国"},
			[]byte{0x00, 0x47, 0x00, 0x72, 0x00, 0x65, 0x00, 0x61, 0x00, 0x74, 0x00, 0x20, 0x4e, 0x2d, 0x56, 0xfd},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := utils.Utf8ToUcs2(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ucs2Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		ucs2 []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"testcase1",
			args{[]byte{0x00, 0x31}},
			"1",
		},
		{
			"testcase2",
			args{[]byte{0x00, 0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00, 0x20, 0x00, 0x77, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6c, 0x00, 0x64}},
			"hello world",
		},
		{"testcase3",
			args{[]byte{0x00, 0x47, 0x00, 0x72, 0x00, 0x65, 0x00, 0x61, 0x00, 0x74, 0x00, 0x20, 0x4e, 0x2d, 0x56, 0xfd}},
			"Great 中国",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := utils.Ucs2ToUtf8(tt.args.ucs2); string(got) != tt.want {
				t.Errorf("Ucs2Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSubmit(t *testing.T) {
	phones := []string{"17011112222", "17500002222"}
	// 160 bytes
	content := "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	mts := cmpp.NewSubmit(ac, phones, content, uint32(codec.B32Seq.NextVal()))
	for _, mt := range mts {
		t.Logf(">>> %v", mt)
		t.Logf("<<< %v", mt.ToResponse(0))
	}

	content2 := content + "hello world"
	mts = cmpp.NewSubmit(ac, phones, content2, uint32(codec.B32Seq.NextVal()))
	for _, mt := range mts {
		t.Logf(">>> %v", mt)
		t.Logf("<<< %v", mt.ToResponse(0))
	}

	content3 := "强大的祖国"
	mts = cmpp.NewSubmit(ac, phones, content3, uint32(codec.B32Seq.NextVal()))
	for _, mt := range mts {
		t.Logf(">>> %v", mt)
		bts := mt.Encode()
		t.Logf("mt.Encode():  %v", bts)
		t.Logf("<<< %v", mt.ToResponse(0))
	}

	content4 := "强大的祖国" + content
	mts = cmpp.NewSubmit(ac, phones, content4, uint32(codec.B32Seq.NextVal()))
	for _, mt := range mts {
		t.Logf(">>> %v", mt)
		t.Logf("<<< %v", mt.ToResponse(0))
	}
}

func TestSubmit_Encode(t *testing.T) {
	phones := []string{"17011112222"}
	mts := cmpp.NewSubmit(ac, phones, Poem, uint32(codec.B32Seq.NextVal()), codec.MtAtTime(time.Now().Add(5*time.Minute)))

	for _, mt := range mts {
		mt := mt.(*cmpp.Submit)
		mt.Version = cmpp.Version(ac.Version)
		t.Logf("mt.String()  : %v", mt)
		resp := mt.ToResponse(0).(*cmpp.SubmitRsp)
		t.Logf("resp.String(): %v", resp)

		enc := mt.Encode()
		t.Logf("Hex MT: %v", enc)
		header := &cmpp.MessageHeader{}
		err := header.Decode(enc[:12])
		if err != nil {
			return
		}
		decMt := &cmpp.Submit{Version: cmpp.Version(ac.Version)}
		err = decMt.Decode(header.SequenceId, enc[12:])
		if err != nil {
			return
		}

		t.Logf("decMt.String()  : %v", decMt)
		bs := decMt.MsgContent()
		content := mt.MsgContent()
		if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
			content = content[6:]
		}
		content, _ = utils.Ucs2ToUtf8(content)
		assert.Equal(t, content, bs)
	}
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
