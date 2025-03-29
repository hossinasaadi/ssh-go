package sshlib

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hossinasaadi/go-socks5"
	"golang.org/x/crypto/ssh"
)

var sshConf *ssh.ClientConfig
var sshConn *ssh.Client
var err error

const timeout time.Duration = 5 * time.Second

func InitSSH(sshAddress string, socks5Address string, sshUser string, sshPass string, privateKeyContent string, remoteAddr string, localAddr string) {
	auths := []ssh.AuthMethod{ssh.Password(sshPass)}

	if privateKeyContent != "" {
		// or get the signer from your private key file directly
		// pemBytes, err := os.ReadFile(privateKeyPath)
		// if err != nil {
		// 	fmt.Println("Reading private key file failed %v", err)
		// 	return
		// }
		pemBytes := []byte(privateKeyContent)
		// create signer
		signer, err := signerFromPem(pemBytes, []byte(sshPass))
		if err != nil {
			fmt.Println("signer failed ", err)
			return
		}

		auths = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	sshConf = &ssh.ClientConfig{
		User:            sshUser,
		Timeout:         timeout,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, err = ssh.Dial("tcp", sshAddress, sshConf)
	if err != nil {
		fmt.Println("error tunnel to server: ", err)
		return
	}
	defer sshConn.Close()

	fmt.Println("connected to ssh server")

	go func() {
		conf := &socks5.Config{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if addr == localAddr {
					return sshConn.DialContext(ctx, network, remoteAddr)
				}
				return sshConn.DialContext(ctx, network, addr)
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
	go handleReconnects(sshAddress)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	return

}

// func handleReconnects(sshAddress string) {
// 	closed := make(chan error, 1)
// 	go func() {
// 		closed <- sshConn.Wait()
// 	}()

//		select {
//		case res := <-closed:
//			println("closed:" + res.Error())
//			sshConn, err = ssh.Dial("tcp", sshAddress, sshConf)
//			if err != nil {
//				println("Failed to reconnect:" + err.Error())
//				return
//			}
//			// Cool we have a new connection, keep going
//			handleReconnects(sshAddress)
//		}
//	}
func handleReconnects(sshAddress string) {
	// Periodically check the connection to google.com
	for {
		err := checkConnection(sshConn)
		if err != nil {
			fmt.Println("SSH tunnel lost or cannot reach google.com:", err)
			sshConn, err = ssh.Dial("tcp", sshAddress, sshConf)
			if err != nil {
				println("Failed to reconnect:" + err.Error())
				return
			}
		} else {
			fmt.Println("SSH tunnel is active and google.com is reachable.")
		}
		time.Sleep(10 * time.Second) // Check every second
	}

}
func checkConnection(sshConn *ssh.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := sshConn.DialContext(ctx, "tcp", "google.com:80")
	if err != nil {
		return err
	}
	conn.Close() // Close the connection after checking
	return nil
}
