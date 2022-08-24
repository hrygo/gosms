package sgip

import (
	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/utils"
	"github.com/hrygo/gosms/utils/snowflake"
)

var conf = []byte(`
{
    "isp": "sgip",
		"loginName": "3037196688",
    "clientId": "3037196688",
    "sharedSecret": "shared secret",
    "version": 18,
    "needReport": 1,
    "smsDisplayNo": "95566",
    "serviceId": "myService",
    "DefaultMsgLevel": 3,
    "feeUserType": 2,
    "FeeTerminalType": 0,
    "feeTerminalId": "",
    "feeType": "05",
    "feeCode": "free",
    "fixedFee": "",
    "LinkId": "",
    "mtValidDuration": 7200000000000,
    "maxConns": 4,
    "mtWindowSize": 16,
    "throughput": 1000
}`)

var ac *codec.AuthConf

func init() {
	ac = codec.Unmarshal(conf)

	codec.B32Seq = utils.NewCycleSequence(1, 7)
	codec.B64Seq = snowflake.NewSnowflake(7, 110)
	codec.BcdSeq = utils.NewBcdSequence("010101")
	sgip.NewSequencer(3037196688, 0)
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

const Poem2 = "Will drink\n" +
	"Don't you see the water of the Yellow River coming up from the sky, rushing to the sea and never returning.\n" +
	"Don't you see the bright mirror of the high hall mourning white hair, like green silk in the morning and snow in the evening.\n" +
	"When you are happy in life, don't make the golden cup empty to the moon.\n" +
	"I'm born to be useful, but I'll come back after all the money is gone.\n" +
	"Cooking sheep and slaughtering cattle is fun, and you will have to drink 300 cups a day.\n" +
	"Master Cen, Dan Qiusheng, don't stop drinking.\n" +
	"Sing a song with you, please listen to it for me.\n" +
	"Bells, drums, and dishes are not expensive. I hope I'll be drunk for a long time and won't wake up.\n" +
	"In ancient times, saints and sages were lonely, and only drinkers kept their names.\n" +
	"The king of Chen used to enjoy banquets and drink ten thousand wine.\n" +
	"Why does the master say less money? He must sell and drink to you.\n" +
	"Five flower horses, thousands of gold fur, hu er will exchange wine, and sell eternal sorrow with you."
