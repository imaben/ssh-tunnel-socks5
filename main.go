package main

import (
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
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

func main() {
	parseConfig()

	sshConf := &ssh.ClientConfig{
		User:            config.Remote.User,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Remote.Passwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	log.Println("connecting", config.Remote.Addr)
	sshConn, err := ssh.Dial("tcp", config.Remote.Addr, sshConf)
	if err != nil {
		log.Fatalln("failed to connect remote server", config.Remote.Addr, err.Error())
	}
	log.Println("connect success")
	defer sshConn.Close()

	socks5Conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshConn.Dial(network, addr)
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
