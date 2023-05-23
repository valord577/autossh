package tunnel2

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"autossh/config"
	"autossh/logs"
)

const (
	autosshFlowBufferSize = "AUTOSSH_FLOW_BUFFER_SIZE"
	defaultFlowBufferSize = 4 * 1024
)

func flowBuffSize() int {
	env := os.Getenv(autosshFlowBufferSize)
	size, _ := strconv.ParseInt(env, 10, 32)
	if size < defaultFlowBufferSize {
		size = defaultFlowBufferSize
	}
	return int(size)
}

var (
	shutdown  bool
	waitgroup sync.WaitGroup
)

func Shutdown() {
	shutdown = true
	waitgroup.Wait()
}

func Startup() error {
	tuns := config.Tunnels()
	if len(tuns) < 1 {
		return errors.New("empty ssh config or tunnels")
	}

	waitgroup.Add(1)
	for _, tun := range tuns {
		go startup(tun)
	}
	return nil
}

func startup(tun *config.Tunnel) {
	defer waitgroup.Done()

	switch tun.ListenOn {
	case config.TunListenOnLocal:
	case config.TunListenOnRemote:
	default:
		return
	}
	listenOn := config.TunListenTypeString(tun.ListenOn)

	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		logs.Warnf("read crypto/rand, err: %s", err.Error())
		return
	}
	uuid := hex.EncodeToString(bs)

	l := logs.With(
		zap.String("uuid", uuid),
		zap.String("service", tun.Service),
		zap.String("listenOn", listenOn),
	)
	l.Infof("start the tunnel service")

	for !shutdown {
		forwarding(tun, l)
	}
}

func forwarding(tun *config.Tunnel, log *logs.Logger) {
	listenOn := tun.ListenOn
	listenAt := tun.ListenAt
	forwardTo := tun.ForwardTo
	sshConfig := tun.SshConfig

	transport := func(src net.Conn, sshClient *ssh.Client) {
		var (
			dst net.Conn
			err error
		)
		switch listenOn {
		case config.TunListenOnLocal:
			dst, err = sshClient.Dial("tcp", forwardTo)
		case config.TunListenOnRemote:
			dst, err = net.DialTimeout("tcp", forwardTo, 10*time.Second)
		}
		if err != nil {
			log.Errorf("target dial, err: %s", err.Error())
			return
		}

		go exflow(dst, src, log)
		go func() {
			e := exflow(src, dst, log)
			if e != nil {
				log.Warnf("exchange flow, addr: '%s', err: %s",
					src.RemoteAddr().String(), e.Error(),
				)
			}
		}()
	}

	log.Infof("ssh dial, alias: %s, address: %s", sshConfig.Alias, sshConfig.Address)
	sshClient, err := ssh.Dial("tcp", sshConfig.Address, sshConfig.Config)
	if err != nil {
		log.Errorf("ssh dial, err: %s", err.Error())
		return
	}
	defer sshClient.Close()

	log.Infof("listen, address: %s", listenAt)
	var listener net.Listener
	switch listenOn {
	case config.TunListenOnLocal:
		listener, err = net.Listen("tcp", listenAt)
	case config.TunListenOnRemote:
		listener, err = sshClient.Listen("tcp", listenAt)
	}
	if err != nil {
		log.Errorf("listen, err: %s", err.Error())
		return
	}
	defer listener.Close()

	done := make(chan struct{}, 1)
	hook := true
	go func() {
		for hook {
			time.Sleep(1 * time.Second)
			if shutdown {
				done <- struct{}{}
				break
			}
		}
	}()
	go func() {
		for {
			conn, e := listener.Accept()
			if e != nil {
				log.Warnf("accept, err: %s", e.Error())
				done <- struct{}{}
				break
			}
			log.Infof("accept: '%s'", conn.RemoteAddr().String())
			go transport(conn, sshClient)
		}
	}()
	<-done
	hook = false
}

func exflow(dst, src net.Conn, log *logs.Logger) error {
	defer func() {
		_ = dst.Close()
		_ = src.Close()
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
			log.Debugf("exchange flow, addr: '%s', written: %d Bytes", srcAddr, nw)
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
	return err
}
