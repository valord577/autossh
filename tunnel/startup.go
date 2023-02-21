package tunnel

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	log "autossh/logger"
)

const (
	autosshFlowBufferSize = "AUTOSSH_FLOW_BUFFER_SIZE"
)

func flowBuffSize() int {
	env := os.Getenv(autosshFlowBufferSize)
	size, _ := strconv.ParseInt(env, 10, 32)
	if size < 2*1024 {
		size = 2 * 1024
	}
	return int(size)
}

type TunServ struct {
	shutdown  bool
	waitgroup sync.WaitGroup
}

func (s *TunServ) Shutdown() {
	s.shutdown = true
	s.waitgroup.Wait()
}

func (s *TunServ) Startup(tunnels []*tunnel) error {
	l := len(tunnels)
	if l < 1 {
		return errors.New("all tunnels don't need to be opened")
	}

	s.waitgroup.Add(l)
	for _, tunnel := range tunnels {
		go s.startupOne(tunnel)
	}
	return nil
}

func (s *TunServ) startupOne(tun *tunnel) {
	defer s.waitgroup.Done()
	if tun == nil {
		return
	}

	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		log.Errorf("read crypto/rand, err: %s", err.Error())
		return
	}
	uuid := base64.URLEncoding.EncodeToString(bs)

	log.Infof("[%s] start the tunnel service: '%s'", uuid, tun.service)
	uuid = uuid + "(" + tun.service + ")"
	s.forwarding(tun.listenOn, uuid, tun.listenAt, tun.forwardTo, tun.sshConfig)
}

func (s *TunServ) forwarding(listenOn listenType, uuid, listenAt, forwardTo string, sshConfig *sshConfig) {
	if listenOn != listenOnLocal &&
		listenOn != listenOnRemote {
		return
	}
	tp := mapListenType(listenOn)

	transport := func(src net.Conn, sshClient *ssh.Client) {
		var (
			dst net.Conn
			err error
		)
		switch listenOn {
		case listenOnLocal:
			dst, err = sshClient.Dial("tcp", forwardTo)
		case listenOnRemote:
			dst, err = net.Dial("tcp", forwardTo)
		}
		if err != nil {
			log.Errorf("[%s] target dial, err: %s", uuid, err.Error())
			return
		}

		wait := &sync.WaitGroup{}
		wait.Add(2)
		go s.exchangeFlow(uuid, dst, src, wait, true)
		go s.exchangeFlow(uuid, src, dst, wait, false)
		wait.Wait()
	}

	for !s.shutdown {
		log.Infof("[%s] forwarding - %s, alias: %s, ssh addr: %s",
			uuid, tp, sshConfig.alias, sshConfig.address)
		sshClient, err := ssh.Dial("tcp", sshConfig.address, sshConfig.config)
		if err != nil {
			log.Errorf("[%s] ssh dial, err: %s", uuid, err.Error())
			continue
		}

		log.Infof("[%s] listen %s, addr: '%s'", uuid, tp, listenAt)
		var listener net.Listener
		switch listenOn {
		case listenOnLocal:
			listener, err = net.Listen("tcp", listenAt)
		case listenOnRemote:
			listener, err = sshClient.Listen("tcp", listenAt)
		}
		if err != nil {
			log.Errorf("[%s] listen %s, err: %s", uuid, tp, err.Error())
			continue
		}

		destroy := make(chan struct{}, 1)
		hook := true
		go func() {
			for hook {
				time.Sleep(1 * time.Second)
				if s.shutdown {
					destroy <- struct{}{}
					break
				}
			}
		}()
		go func() {
			for {
				conn, e := listener.Accept()
				if e != nil {
					log.Warnf("[%s] %s accept, err: %s", uuid, tp, e.Error())
					destroy <- struct{}{}
					break
				}
				log.Infof("[%s] %s accept: '%s'", uuid, tp, conn.RemoteAddr().String())
				go transport(conn, sshClient)
			}
		}()
		<-destroy
		hook = false

		_ = listener.Close()
		_ = sshClient.Close()
	}
}

func (s *TunServ) exchangeFlow(uuid string, dst, src net.Conn, wait *sync.WaitGroup, logs bool) {
	defer func() {
		_ = dst.Close()
		_ = src.Close()
		wait.Done()
	}()

	var (
		err error
		buf = make([]byte, flowBuffSize())

		srcAddr = src.RemoteAddr().String()
	)
	for !s.shutdown {
		nr, er := src.Read(buf)
		if er != nil {
			err = er
			break
		}
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			if logs {
				log.Debugf("[%s] exchange flow, addr: '%s', written: %d Bytes", uuid, srcAddr, nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = errors.New("short write")
				break
			}
		}
	}
	if logs && err != nil {
		log.Warnf("[%s] exchange flow, addr: '%s', err: %s", uuid, srcAddr, err.Error())
	}
}
