package tunnel

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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

var (
	shutdown = false

	wg sync.WaitGroup
)

func startup(tunnels []*tunnel) error {
	l := len(tunnels)
	if l < 1 {
		return errors.New("all tunnels don't need to be opened")
	}

	wg.Add(l)
	for _, tunnel := range tunnels {
		go startupOne(tunnel)
	}
	// block and listen for signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	s := <-sig
	log.Infof("recv signal: %d", s)

	shutdown = true
	wg.Wait()
	return nil
}

func startupOne(tun *tunnel) {
	defer wg.Done()
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

	switch tun.listenOn {
	case listenOnLocal:
		listenLocal(uuid, tun.listenAt, tun.forwardTo, tun.sshConfig)
	case listenOnRemote:
		listenRemote(uuid, tun.listenAt, tun.forwardTo, tun.sshConfig)
	default:
	}
}

func shutdownHook(destroy chan struct{}) {
	for {
		time.Sleep(1 * time.Second)
		if shutdown {
			destroy <- struct{}{}
			break
		}
	}
}

func listenLocal(uuid, listenAt, forwardTo string, sshConfig *sshConfig) {
	transport := func(local net.Conn, sshClient *ssh.Client) {
		log.Infof("[%s] target dial: %s", uuid, forwardTo)
		target, err := sshClient.Dial("tcp", forwardTo)
		if err != nil {
			log.Errorf("[%s] target dial, err: %s", uuid, err.Error())
			return
		}

		wait := &sync.WaitGroup{}
		wait.Add(2)
		go exchangeFlow(uuid, target, local, wait, true)
		go exchangeFlow(uuid, local, target, wait, false)
		wait.Wait()
	}

	for !shutdown {
		log.Infof("[%s] forwarding - local, alias: %s, ssh addr: %s", uuid, sshConfig.alias, sshConfig.address)
		sshClient, err := ssh.Dial("tcp", sshConfig.address, sshConfig.config)
		if err != nil {
			log.Errorf("[%s] ssh dial, err: %s", uuid, err.Error())
			return
		}

		log.Infof("[%s] listen local, addr: '%s'", uuid, listenAt)
		listener, err := net.Listen("tcp", listenAt)
		if err != nil {
			log.Errorf("[%s] listen local, err: %s", uuid, err.Error())
			continue
		}

		destroy := make(chan struct{}, 1)
		go shutdownHook(destroy)
		go func() {
			for {
				conn, e := listener.Accept()
				if e != nil {
					log.Warnf("[%s] local accept, err: %s", uuid, e.Error())
					destroy <- struct{}{}
					break
				}
				log.Infof("[%s] local accept: '%s'", uuid, conn.RemoteAddr().String())
				go transport(conn, sshClient)
			}
		}()
		<-destroy

		_ = listener.Close()
		_ = sshClient.Close()
	}
}

func listenRemote(uuid, listenAt, forwardTo string, sshConfig *sshConfig) {
	transport := func(remote net.Conn) {
		target, err := net.Dial("tcp", forwardTo)
		if err != nil {
			log.Errorf("[%s] target dial, err: %s", uuid, err.Error())
			return
		}
		log.Infof("[%s] target dial: '%s'", uuid, target.RemoteAddr().String())

		wait := &sync.WaitGroup{}
		wait.Add(2)
		go exchangeFlow(uuid, target, remote, wait, true)
		go exchangeFlow(uuid, remote, target, wait, false)
		wait.Wait()
	}

	for !shutdown {
		log.Infof("[%s] forwarding - remote, alias: %s, ssh addr: %s", uuid, sshConfig.alias, sshConfig.address)
		sshClient, err := ssh.Dial("tcp", sshConfig.address, sshConfig.config)
		if err != nil {
			log.Errorf("[%s] ssh dial, err: %s", uuid, err.Error())
			continue
		}

		log.Infof("[%s] listen remote, addr: '%s'", uuid, listenAt)
		listener, err := sshClient.Listen("tcp", listenAt)
		if err != nil {
			log.Errorf("[%s] listen remote, err: %s", uuid, err.Error())
			continue
		}

		destroy := make(chan struct{}, 1)
		go shutdownHook(destroy)
		go func() {
			for {
				conn, e := listener.Accept()
				if e != nil {
					log.Warnf("[%s] remote accept, err: %s", uuid, e.Error())
					destroy <- struct{}{}
					break
				}
				log.Infof("[%s] remote accept: '%s'", uuid, conn.RemoteAddr().String())
				go transport(conn)
			}
		}()
		<-destroy

		_ = listener.Close()
		_ = sshClient.Close()
	}
}

func exchangeFlow(uuid string, dst, src net.Conn, wait *sync.WaitGroup, logs bool) {
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
	for !shutdown {
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
