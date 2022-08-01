package smgp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

const (
	LoginLen     = 42
	LoginRespLen = 33
)

type Login struct {
	MessageHeader               //  【12字节】消息头
	clientID            string  //  【8字节】客户端用来登录服务器端的用户账号。
	authenticatorClient []byte  //  【16字节】客户端认证码，用来鉴别客户端的合法性。
	loginMode           byte    //  【1字节】客户端用来登录服务器端的登录类型。
	timestamp           uint32  //  【4字节】时间戳
	Version             Version //  【1字节】客户端支持的协议版本号

	// 非协议内容，调用ToResponse前需设置
	secret string
}

type LoginRsp struct {
	MessageHeader               // 协议头, 12字节
	status              Status  // 状态码，4字节
	authenticatorServer []byte  // 认证串，16字节
	Version             Version // 版本，1字节
}

func NewLogin(cl *auth.Client, seq uint32) *Login {
	lo := &Login{}
	lo.PacketLength = LoginLen
	lo.RequestId = SMGP_LOGIN
	lo.SequenceId = seq
	lo.clientID = cl.ClientId
	lo.loginMode = 2
	var ts string
	ts, lo.timestamp = utils.Now()
	authMd5 := md5.Sum(bytes.Join([][]byte{
		[]byte(cl.ClientId),
		make([]byte, 7),
		[]byte(cl.SharedSecret),
		[]byte(ts),
	}, nil))
	lo.authenticatorClient = authMd5[:]
	lo.Version = Version(cl.Version)
	return lo
}

func (l *Login) Encode() []byte {
	frame := l.MessageHeader.Encode()
	if len(frame) == LoginLen && l.PacketLength == LoginLen {
		copy(frame[12:20], l.clientID)
		copy(frame[20:36], l.authenticatorClient)
		frame[36] = l.loginMode
		binary.BigEndian.PutUint32(frame[37:41], l.timestamp)
		frame[41] = byte(l.Version)
	}
	return frame
}

func (l *Login) Decode(seq uint32, frame []byte) error {
	l.PacketLength = codec.HeadLen + uint32(len(frame))
	l.RequestId = SMGP_LOGIN
	l.SequenceId = seq
	l.clientID = string(frame[0:8])
	l.authenticatorClient = frame[8:24]
	l.loginMode = frame[24]
	l.timestamp = binary.BigEndian.Uint32(frame[25:29])
	l.Version = Version(frame[29])
	return nil
}

func (l *Login) String() string {
	return fmt.Sprintf("{Header: %s, clientID: %s, authenticatorClient: %x, logoinMode: %x, timestamp: %010d, version: %s}",
		&l.MessageHeader, l.clientID, l.authenticatorClient, l.loginMode, l.timestamp, l.Version)
}

func (l *Login) Check(cli *auth.Client) Status {
	// 大版本不匹配
	if !l.Version.MajorMatch(cli.Version) {
		return Status(22)
	}

	authSource := l.authenticatorClient
	authMd5 := md5.Sum(bytes.Join([][]byte{
		[]byte(cli.ClientId),
		make([]byte, 7),
		[]byte(cli.SharedSecret),
		[]byte(utils.TimeStamp2Str(l.timestamp)),
	}, nil))
	log.Debugf("[AuthCheck] input  : %x", authSource)
	log.Debugf("[AuthCheck] compute: %x", authMd5)
	ok := bytes.Equal(authSource, authMd5[:])
	// 配置不做校验或校验通过时返回0
	if ok {
		l.SetSecret(cli.SharedSecret)
		return Status(0)
	}
	return Status(21)
}

func (l *Login) ToResponse(code uint32) codec.Pdu {
	rsp := &LoginRsp{}
	rsp.PacketLength = LoginRespLen
	rsp.RequestId = SMGP_LOGIN_RESP
	rsp.SequenceId = l.SequenceId
	rsp.status = Status(code)
	if rsp.status == Status(0) {
		md5Auth := md5.Sum(bytes.Join([][]byte{
			{byte(code)},
			[]byte(l.clientID),
			[]byte(l.secret),
		}, nil))
		rsp.authenticatorServer = md5Auth[:]
	} else {
		rsp.authenticatorServer = make([]byte, 16, 16)
	}
	rsp.Version = l.Version
	return rsp
}

func (l *Login) SetSecret(secret string) {
	l.secret = secret
}

func (l *Login) Log() []log.Field {
	ls := l.MessageHeader.Log()
	return append(ls,
		log.String("clientID", l.clientID),
		log.String("authenticatorSource", hex.EncodeToString(l.authenticatorClient)),
		log.Int8("loginMode", int8(l.loginMode)),
		log.String("timestamp", fmt.Sprintf("%010d", l.timestamp)),
		log.String("version", hex.EncodeToString([]byte{byte(l.Version)})),
	)
}

func (r *LoginRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	var index int
	if len(frame) == int(r.PacketLength) {
		index = 12
		binary.BigEndian.PutUint32(frame[index:index+4], uint32(r.status))
		index += 4
		copy(frame[index:index+16], r.authenticatorServer)
		index += 16
		frame[index] = byte(r.Version)
	}
	return frame
}

func (r *LoginRsp) Decode(seq uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.RequestId = SMGP_LOGIN_RESP
	r.SequenceId = seq
	var index int
	r.status = Status(binary.BigEndian.Uint32(frame[0 : index+4]))
	index = 4
	r.authenticatorServer = frame[index : index+16]
	index += 16
	r.Version = Version(frame[index])
	return nil
}

func (r *LoginRsp) String() string {
	return fmt.Sprintf("{ Header: %s, status: \"%s\", authenticatorISMG: %x, version: %s }",
		&r.MessageHeader, r.status, r.authenticatorServer, r.Version)
}

func (r *LoginRsp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls,
		log.String("status", r.status.String()),
		log.String("authenticatorISMG", hex.EncodeToString(r.authenticatorServer)),
		log.String("version", hex.EncodeToString([]byte{byte(r.Version)})),
	)
}

func (l *Login) ClientID() string {
	return l.clientID
}

func (l *Login) AuthenticatorClient() []byte {
	return l.authenticatorClient
}

func (l *Login) LoginMode() byte {
	return l.loginMode
}

func (l *Login) Timestamp() uint32 {
	return l.timestamp
}

func (r *LoginRsp) AuthenticatorServer() []byte {
	return r.authenticatorServer
}

func (r *LoginRsp) Status() Status {
	return r.status
}
