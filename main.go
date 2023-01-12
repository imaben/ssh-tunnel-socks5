package main

import (
	"context"
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"sync"
	"time"
)

type cfgRemote struct {
	Addr   string
	User   string
	Passwd string
}

type cfgLocal struct {
	Listen string
}

type Config struct {
	Remote cfgRemote
	Local  cfgLocal
}

var config Config

func parseConfig() {
	cfg := flag.String("c", "config.toml", "config file")
	flag.Parse()
	if _, err := toml.DecodeFile(*cfg, &config); err != nil {
		log.Fatal(err.Error())
	}
}

type RemoteProxy struct {
	mutex     sync.Mutex
	cfg       *ssh.ClientConfig
	clt       *ssh.Client
	sleepTime time.Duration
}

func NewRemoteProxy() *RemoteProxy {
	cfg := &ssh.ClientConfig{
		User:            config.Remote.User,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Remote.Passwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	rp := &RemoteProxy{
		cfg:       cfg,
		sleepTime: 1,
	}
	rp.connect()
	go rp.daemon()
	return rp
}
func (rp *RemoteProxy) connect() {
	var err error
	log.Println("connecting", config.Remote.Addr)
	rp.clt, err = ssh.Dial("tcp", config.Remote.Addr, rp.cfg)
	if err != nil {
		log.Println("failed to connect remote server", config.Remote.Addr, err.Error())
	}
	log.Println("connect success")
}

func (rp *RemoteProxy) daemon() {
	for {
		if rp.clt != nil {
			rp.clt.Wait()
		}
		log.Println("remote connection has been disconnected, retry...")
		rp.sleep()
		rp.connect()
	}
}

func (rp *RemoteProxy) sleep() {
	log.Println("sleep")
	tm := rp.sleepTime
	if tm > 60 {
		tm = 60
	}
	time.Sleep(time.Second * tm)
	rp.sleepTime += 1
}

func (rp *RemoteProxy) Dial(n, addr string) (net.Conn, error) {
	clt := rp.clt
	if clt == nil {
		log.Println("client is not ready")
		return nil, errors.New("client is nil")
	}
	return clt.Dial(n, addr)
}

func (rp *RemoteProxy) Destroy() {
	rp.clt.Close()
}

func main() {
	parseConfig()

	rp := NewRemoteProxy()
	defer rp.Destroy()
	socks5Conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return rp.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(socks5Conf)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	log.Println("start listen", config.Local.Listen)
	if err := serverSocks.ListenAndServe("tcp", config.Local.Listen); err != nil {
		log.Fatalln("failed to create socks5 server", err.Error())
	}
}
