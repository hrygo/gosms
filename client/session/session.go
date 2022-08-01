package session

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/codec/cmpp"
	"github.com/hrygo/gosmsn/codec/smgp"
)

type Session struct {
	sync.Mutex

	con           net.Conn
	cli           *auth.Client
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
func NewSession(isp string, cli *auth.Client, con net.Conn) *Session {
	sc := &Session{con: con, cli: cli, serverName: isp, stat: StatConnect}
	// _ = con.SetDeadline(time.Now().Add(time.Second))
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

func (s *Session) HealthCheck() bool {
	ok := s != nil && s.stat == StatLogin && s.con != nil && s.cli != nil && s.cancel != nil

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
	// TODO  通过心跳验证连接是否正常
	return nil
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
	defer close(s.cancel)

	// 关闭连接
	s.stat = StatClosing
	_ = s.con.Close()
	// 发送取消信号
	s.cancel <- struct{}{}
	s.con = nil
	s.cli = nil
}

// startReceiver 启动接收服务
func (s *Session) startReceiver() {
	go func() {
		for {
			select {
			case <-s.cancel:
				log.Warn("Receive cancel signal, Receiver exit!")
				return
			default:
				if !s.HealthCheck() {
					log.Warn("Session status incorrect, Receiver exit!")
					return
				}
				// TODO receive
				time.Sleep(time.Millisecond)
			}
		}
	}()
}

func (s *Session) login() error {
	var pdu codec.Pdu
	var respLen = 27
	switch s.serverName {
	case SGIP:
		{
		}
	case CMPP:
		{
			pdu = cmpp.NewConnect(s.cli, uint32(codec.B32Seq.NextVal()))
			if cmpp.V30.MajorMatch(s.cli.Version) {
				respLen = cmpp.ConnectRspPktLenV3
			} else {
				respLen = cmpp.ConnectRspPktLenV2
			}
		}
	case SMGP:
		{
			pdu = smgp.NewLogin(s.cli, uint32(codec.B32Seq.NextVal()))
			respLen = smgp.LoginRespLen
		}
	}

	_, err := s.con.Write(pdu.Encode())
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Send login to %v.", s.con.RemoteAddr()), pdu.Log()...)

	data := make([]byte, respLen)
	n, err := s.con.Read(data[:])
	if err != nil || n != respLen {
		return err
	}

	var pkl = binary.BigEndian.Uint32(data[:4])
	if pkl != uint32(n) {
		return errors.New(fmt.Sprintf("packet length error: expect %d, got %d", n, pkl))
	}
	var cmd = binary.BigEndian.Uint32(data[4:8])
	var seq = binary.BigEndian.Uint32(data[8:12])

	switch s.serverName {
	case SGIP:
		{
		}
	case CMPP:
		{
			if cmd != uint32(cmpp.CMPP_CONNECT_RESP) {
				return errors.New(fmt.Sprintf("CommandId error: expect %x, got %x", cmpp.CMPP_CONNECT_RESP, cmd))
			}
			resp := &cmpp.ConnectResp{Version: cmpp.Version(s.cli.Version)}
			err := resp.Decode(seq, data[12:])
			if err != nil {
				return err
			}
			if resp.Status() != cmpp.ConnStatusOK {
				return errors.New(fmt.Sprintf("Login error with return \"%s\"", resp.Status().String()))
			}
			log.Info("Login result", resp.Log()...)
		}
	case SMGP:
		{
			if cmd != uint32(smgp.SMGP_LOGIN_RESP) {
				return errors.New(fmt.Sprintf("CommandId error: expect %x, got %x", smgp.SMGP_LOGIN_RESP, cmd))
			}
			resp := &smgp.LoginRsp{Version: smgp.Version(s.cli.Version)}
			err := resp.Decode(seq, data[12:])
			if err != nil {
				return err
			}
			if resp.Status() != smgp.Status(0) {
				return errors.New(fmt.Sprintf("Login error with return \"%s\"", resp.Status().String()))
			}
			log.Info("Login result", resp.Log()...)
		}
	}
	s.stat = StatLogin
	s.activeTime = time.Now()
	return nil
}
