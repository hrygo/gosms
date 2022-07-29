package cmpp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

const ConnectPktLen = 12 + 6 + 16 + 1 + 4

type Connect struct {
	MessageHeader               // +12 = 12：消息头
	sourceAddr          string  // +6 = 18：源地址，此处为 SP_Id
	authenticatorSource []byte  // +16 = 34： 用于鉴别源地址。其值通过单向 MD5 hash 计算得出，表示如下: authenticatorSource = MD5(Source_Addr+9 字节的 0 +shared secret+timestamp) Shared secret 由中国移动与源地址实 体事先商定，timestamp 格式为: MMDDHHMMSS，即月日时分秒，10 位。
	version             Version // +1 = 35：双方协商的版本号(高位 4bit 表示主 版本号,低位 4bit 表示次版本号)，对 于3.0的版本，高4bit为3，低4位为 0
	timestamp           uint32  // +4 = 39：时间戳的明文,由客户端产生,格式为 MMDDHHMMSS，即月日时分秒，10 位数字的整型，右对齐。

	// 非协议内容，调用ToResponse前需设置
	secret string
}

func (c *Connect) SourceAddr() string {
	return c.sourceAddr
}
func (c *Connect) AuthenticatorSource() []byte {
	return c.authenticatorSource
}

func (c *Connect) Timestamp() uint32 {
	return c.timestamp
}

func (c *Connect) Version() Version {
	return c.version
}

func (c *Connect) SetSecret(secret string) {
	c.secret = secret
}

const ConnectRspPktLenV3 = 12 + 4 + 16 + 1
const ConnectRspPktLenV2 = 12 + 1 + 16 + 1

type ConnectResp struct {
	MessageHeader                // 协议头, 12字节
	status            ConnStatus // 状态码，3.0版本4字节，2.0版本1字节
	authenticatorISMG []byte     // 认证串，16字节
	version           Version    // 版本，1字节
}

func (r *ConnectResp) AuthenticatorISMG() []byte {
	return r.authenticatorISMG
}

func (r *ConnectResp) Version() Version {
	return r.version
}

func (r *ConnectResp) Status() ConnStatus {
	return r.status
}

func NewConnect(cl *client.Client, seq uint32) *Connect {
	con := &Connect{}
	con.TotalLength = ConnectPktLen
	con.CommandId = CMPP_CONNECT
	con.SequenceId = seq
	con.version = Version(cl.Version)
	con.sourceAddr = cl.ClientId
	var ts string
	ts, con.timestamp = utils.Now()
	authMd5 := md5.Sum(bytes.Join([][]byte{
		[]byte(cl.ClientId),
		make([]byte, 9),
		[]byte(cl.SharedSecret),
		[]byte(ts),
	}, nil))
	con.authenticatorSource = authMd5[:]
	return con
}

func (c *Connect) Encode() []byte {
	frame := c.MessageHeader.Encode()
	if len(frame) == ConnectPktLen && c.TotalLength == ConnectPktLen {
		copy(frame[12:18], c.sourceAddr)
		copy(frame[18:34], c.authenticatorSource)
		frame[34] = byte(c.version)
		binary.BigEndian.PutUint32(frame[35:39], c.timestamp)
		return frame
	}
	return nil
}

func (c *Connect) Decode(seq uint32, frame []byte) error {
	c.TotalLength = ConnectPktLen
	c.CommandId = CMPP_CONNECT
	c.SequenceId = seq
	c.sourceAddr = utils.TrimStr(frame[0:6])
	c.authenticatorSource = frame[6:22]
	c.version = Version(frame[22])
	c.timestamp = binary.BigEndian.Uint32(frame[23:27])
	return nil
}

func (c *Connect) Log() []log.Field {
	ls := c.MessageHeader.Log()
	ls = append(ls,
		log.String("clientID", c.sourceAddr),
		log.String("authenticatorSource", hex.EncodeToString(c.authenticatorSource)),
		log.String("version", hex.EncodeToString([]byte{byte(c.version)})),
		log.String("timestamp", fmt.Sprintf("%010d", c.timestamp)))
	return ls
}

func (c *Connect) Check(cli *client.Client) ConnStatus {
	if cli == nil {
		return ConnStatusInvalidSrcAddr
	}
	if !c.version.MajorMatch(cli.Version) {
		return ConnStatusVerTooHigh
	}

	authSource := c.authenticatorSource
	authMd5 := md5.Sum(bytes.Join([][]byte{
		[]byte(cli.ClientId),
		make([]byte, 9),
		[]byte(cli.SharedSecret),
		[]byte(utils.TimeStamp2Str(c.timestamp)),
	}, nil))
	log.Debugf("[AuthCheck] input  : %x", authSource)
	log.Debugf("[AuthCheck] compute: %x", authMd5)
	ok := bytes.Equal(authSource, authMd5[:])
	if ok {
		c.SetSecret(cli.SharedSecret)
		return ConnStatusOK
	}
	return ConnStatusAuthFailed
}

func (c *Connect) ToResponse(code uint32) codec.Pdu {
	rsp := &ConnectResp{}
	// 3.x 与 2.x Status长度不同
	if V30.MajorMatchV(c.version) {
		rsp.TotalLength = ConnectRspPktLenV3
	} else {
		rsp.TotalLength = ConnectRspPktLenV2
	}
	rsp.CommandId = CMPP_CONNECT_RESP
	rsp.SequenceId = c.SequenceId
	rsp.status = ConnStatus(code)

	var bs []byte
	if V30.MajorMatchV(c.version) {
		bs = []byte{0, 0, 0, byte(rsp.status)}
	} else {
		bs = []byte{byte(rsp.status)}
	}
	if rsp.status == ConnStatusOK {
		md5Auth := md5.Sum(bytes.Join([][]byte{
			bs,
			[]byte(c.sourceAddr),
			[]byte(c.secret),
		}, nil))
		rsp.authenticatorISMG = md5Auth[:]
	} else {
		rsp.authenticatorISMG = make([]byte, 16, 16)
	}
	rsp.version = c.version
	return rsp
}

// 以下为Response

func (r *ConnectResp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	var index int
	if len(frame) == int(r.TotalLength) {
		index = 12
		if V30.MajorMatchV(r.version) {
			binary.BigEndian.PutUint32(frame[index:index+4], uint32(r.status))
			index += 4
		} else {
			frame[index] = byte(r.status)
			index++
		}
		copy(frame[index:index+16], r.authenticatorISMG)
		index += 16
		frame[index] = byte(r.version)
	}
	return frame
}

func (r *ConnectResp) Decode(seq uint32, frame []byte) error {
	if V30.MajorMatchV(r.version) {
		r.TotalLength = ConnectRspPktLenV3
	} else {
		r.TotalLength = ConnectRspPktLenV2
	}
	r.CommandId = CMPP_CONNECT_RESP
	r.SequenceId = seq

	var index int
	if V30.MajorMatchV(r.version) {
		index += 3
	}
	r.status = ConnStatus(frame[index])
	index += 1
	r.authenticatorISMG = frame[index : index+16]
	index += 16
	r.version = Version(frame[index])
	return nil
}

func (r *ConnectResp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	ls = append(ls,
		log.String("status", r.status.String()),
		log.String("authenticatorISMG", hex.EncodeToString(r.authenticatorISMG)),
		log.String("version", hex.EncodeToString([]byte{byte(r.version)})))
	return ls
}

type ConnStatus byte

const (
	ConnStatusOK ConnStatus = iota
	ConnStatusInvalidStruct
	ConnStatusInvalidSrcAddr
	ConnStatusAuthFailed
	ConnStatusVerTooHigh
	ConnStatusOthers
)

func (i ConnStatus) String() string {
	return fmt.Sprintf("%d: %s", i, ConnectStatusMap[uint32(i)])
}

var ConnectStatusMap = map[uint32]string{
	0: "成功",
	1: "消息结构错",
	2: "非法源地址",
	3: "认证错",
	4: "版本太高",
	5: "其他错误",
}
