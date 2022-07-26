package session

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/cmpp"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/codec/smgp"
)

type Session struct {
	sync.Mutex

	con           net.Conn
	authConf      *codec.AuthConf
	serverName    string
	stat          stat
	counter       uint64
	periodCounter uint64 // 短周期内的计数器，用于LRU排序算法
	createTime    time.Time
	activeTime    time.Time
	cancel        chan struct{} // 用以接收停止信号
}

// 会话状态
type stat byte

const (
	StatConnect stat = iota
	StatLogin
	StatClosing

	CMPP = "cmpp"
	SGIP = "sgip"
	SMGP = "smgp"
)

// NewSession 创建一个新会话并登录，且启动定时器和接收服务
func NewSession(isp string, ac *codec.AuthConf, con net.Conn) *Session {
	sc := &Session{con: con, authConf: ac, serverName: isp, stat: StatConnect}
	err := sc.login()
	if err != nil {
		log.Error("create session error: " + err.Error())
		return nil
	}
	sc.cancel = make(chan struct{}, 1)
	sc.startReceiver()
	sc.createTime = time.Now()
	sc.activeTime = time.Now()
	return sc
}

func (s *Session) ResetCounter() {
	s.periodCounter = 0
}

func (s *Session) AddCounter() {
	s.AddCounterN(1)
}

func (s *Session) AddCounterN(n uint64) {
	s.Lock()
	defer s.Unlock()
	s.counter += n
	s.periodCounter += n
	s.activeTime = time.Now()
}

// LruPriority 最近最少使用的优先级最高
func (s *Session) LruPriority() uint64 {
	// 非正常会话优先级最低
	if !s.HealthCheck() {
		return 0
	}
	return 1e10 - s.periodCounter
}

func (s *Session) Close() {
	if s == nil || s.stat == StatClosing {
		return
	}

	s.Lock()
	defer s.Unlock()

	// 关闭连接
	s.stat = StatClosing
	_ = s.con.Close()
	// 发送取消信号
	s.cancel <- struct{}{}
	s.con = nil
	s.authConf = nil
}

// startReceiver 启动接收服务
func (s *Session) startReceiver() {
	go func() {
		// 无限循环接收
		for {
			// 获取取消信号，如果通道阻塞则执行正常业务逻辑，否则取消循环，停止当前goroutine
			select {
			case <-s.cancel:
				log.Warnf("[%s] Receive cancel signal, Receiver exit!", s.serverName)
				return
			default:
				{
					if !s.HealthCheck() {
						log.Warnf("[%s] Session status incorrect, Receiver exit!", s.serverName)
						return
					}
					head := make([]byte, 12)
					n, err := s.con.Read(head)
					if n != 12 || err != nil {
						log.Errorf("[%s] Read Packet Head error: %v. Session closing...", s.serverName, err)
						s.Close()
						return
					}
					log.Debugf("[%s] Read Packet Head: %x", s.serverName, head)
					pkl, cmd, seq := codec.UnpackHead(head)
					// 报文长度检查
					if pkl > codec.PacketMax || pkl < codec.HeadLen {
						log.Errorf("[%s] Read Packet Length(%d) abnormal, Session closing...", s.serverName, pkl)
						s.Close()
						return
					}
					// 其他方面，客户端，弱化了报文合法性检查，后续不合法报文将直接丢弃

					var buff []byte
					var bodyLen = int(pkl) - 12
					if bodyLen > 0 {
						buff = make([]byte, bodyLen)
						n, err = s.con.Read(buff)
						if n != bodyLen || err != nil {
							log.Errorf("[%s] Read Packet Body error: %v. Session closing...", s.serverName, err)
							s.Close()
							return
						}
					}
					log.Debugf("[%s] Read Packet Body: %x", s.serverName, buff)
					s.onTraffic(cmd, seq, buff)
				}
			}
		}
		// 结束无限循环
	}()
}

func (s *Session) onTraffic(cmd, seq uint32, buff []byte) {
	switch s.serverName {
	case CMPP:
		s.onTrafficCmpp(cmd, seq, buff)
	case SMGP:
		s.onTrafficSmgp(cmd, seq, buff)
	case SGIP:
		s.onTrafficSgip(cmd, seq, buff)
	}
}

func (s *Session) login() error {
	var pdu codec.Pdu
	var respLen = 27
	switch s.serverName {
	case SGIP:
		{
			pdu = sgip.NewBind(s.authConf, 1)
			respLen = sgip.BindRspPkgLen
		}
	case CMPP:
		{
			pdu = cmpp.NewConnect(s.authConf, uint32(codec.B32Seq.NextVal()))
			if cmpp.V30.MajorMatch(s.authConf.Version) {
				respLen = cmpp.ConnectRspPktLenV3
			} else {
				respLen = cmpp.ConnectRspPktLenV2
			}
		}
	case SMGP:
		{
			pdu = smgp.NewLogin(s.authConf, uint32(codec.B32Seq.NextVal()))
			respLen = smgp.LoginRespLen
		}
	}

	_, err := s.con.Write(pdu.Encode())
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("[%s] Send login to %v.", s.serverName, s.con.RemoteAddr()), pdu.Log()...)

	data := make([]byte, respLen)
	n, err := s.con.Read(data[:])
	if err != nil || n != respLen {
		return err
	}

	var pkl = binary.BigEndian.Uint32(data[:4])
	if pkl != uint32(n) {
		return errors.New(fmt.Sprintf("[%s] Packet length error: expect %d, got %d", s.serverName, n, pkl))
	}
	var cmd = binary.BigEndian.Uint32(data[4:8])
	var seq = binary.BigEndian.Uint32(data[8:12])

	switch s.serverName {
	case SGIP:
		{
			if cmd != uint32(sgip.SGIP_BIND_RESP) {
				return errors.New(fmt.Sprintf("[%s] CommandId error: expect %x, got %x", s.serverName, sgip.SGIP_BIND_RESP, cmd))
			}
			resp := &sgip.BindRsp{}
			err := resp.Decode(seq, data[12:])
			if err != nil {
				return err
			}
			if resp.Status != sgip.Status(0) {
				return errors.New(fmt.Sprintf("[%s] Login error with return \"%s\"", s.serverName, resp.Status.String()))
			}
			log.Info(fmt.Sprintf("[%s] Login result", s.serverName), resp.Log()...)
		}
	case CMPP:
		{
			if cmd != uint32(cmpp.CMPP_CONNECT_RESP) {
				return errors.New(fmt.Sprintf("[%s] CommandId error: expect %x, got %x", s.serverName, cmpp.CMPP_CONNECT_RESP, cmd))
			}
			resp := &cmpp.ConnectResp{Version: cmpp.Version(s.authConf.Version)}
			err := resp.Decode(seq, data[12:])
			if err != nil {
				return err
			}
			if resp.Status() != cmpp.ConnStatusOK {
				return errors.New(fmt.Sprintf("[%s] Login error with return \"%s\"", s.serverName, resp.Status().String()))
			}
			log.Info(fmt.Sprintf("[%s] Login result", s.serverName), resp.Log()...)
		}
	case SMGP:
		{
			if cmd != uint32(smgp.SMGP_LOGIN_RESP) {
				return errors.New(fmt.Sprintf("[%s] CommandId error: expect %x, got %x", s.serverName, smgp.SMGP_LOGIN_RESP, cmd))
			}
			resp := &smgp.LoginRsp{Version: smgp.Version(s.authConf.Version)}
			err := resp.Decode(seq, data[12:])
			if err != nil {
				return err
			}
			if resp.Status() != smgp.Status(0) {
				return errors.New(fmt.Sprintf("[%s] Login error with return \"%s\"", s.serverName, resp.Status().String()))
			}
			log.Info(fmt.Sprintf("[%s] Login result", s.serverName), resp.Log()...)
		}
	}
	s.stat = StatLogin
	s.activeTime = time.Now()
	return nil
}

func (s *Session) HealthCheck() bool {
	ok := s != nil && s.stat == StatLogin && s.con != nil && s.authConf != nil && s.cancel != nil

	// 活跃状态为1分钟前，则发送心跳验证
	if ok && s.activeTime.Add(time.Minute).Before(time.Now()) {
		err := s.ActiveTest()
		if err != nil {
			return false
		}
	}
	return ok
}

func (s *Session) ActiveTest() error {
	msg := fmt.Sprintf("[%s] ActiveTest", s.serverName)
	var active codec.RequestPdu
	switch s.serverName {
	case CMPP:
		active = cmpp.NewActiveTest(uint32(codec.B32Seq.NextVal()))
	case SGIP:
		// active =
	case SMGP:
		active = smgp.NewActiveTest(uint32(codec.B32Seq.NextVal()))
	}
	if active == nil {
		return nil
	}
	pack := active.Encode()
	_, err := s.con.Write(pack)
	if err == nil {
		log.Info(msg, active.Log()...)
	} else {
		log.Error(msg + err.Error())
		return nil
	}
	return nil
}
