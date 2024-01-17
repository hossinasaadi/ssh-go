package sshlib

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/hossinasaadi/go-socks5"
	"golang.org/x/crypto/ssh"
)

func InitSSH(sshAddress string, socks5Address string, sshUser string, sshPass string) {
	sshConf := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            []ssh.AuthMethod{ssh.Password(sshPass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, err := ssh.Dial("tcp", sshAddress, sshConf)
	if err != nil {
		fmt.Println("error tunnel to server: ", err)
		return
	}
	defer sshConn.Close()

	fmt.Println("connected to ssh server")

	go func() {
		conf := &socks5.Config{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sshConn.Dial(network, addr)
			},
			DisableFQDN: true,
		}

		serverSocks, err := socks5.New(conf)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := serverSocks.ListenAndServe("tcp", socks5Address); err != nil {
			fmt.Println("failed to create socks5 server", err)
		}
		if err := serverSocks.ListenAndServe("udp", socks5Address); err != nil {
			fmt.Println("failed to create socks5 server", err)
		}

	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	return

}
