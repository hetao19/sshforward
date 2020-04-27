package g

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type sshSession struct {
	listenAddr string
	remoteAddr string
	sshClient  *ssh.Client
}

func NewsshSession(listenAddr, remoteAddr string, sshClient *ssh.Client) *sshSession {
	return &sshSession{
		listenAddr: listenAddr,
		remoteAddr: remoteAddr,
		sshClient:  sshClient,
	}
}

func (s *sshSession) handleConn(conn net.Conn) {
	log.Printf("accept %s\n", conn.RemoteAddr())
	remote, err := s.sshClient.Dial("tcp", s.remoteAddr)
	if err != nil {
		log.Printf("dial %s error", s.remoteAddr)
		return
	}
	log.Printf("%s --> %s connected.\n", conn.RemoteAddr(), s.remoteAddr)
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		io.Copy(remote, conn)
		remote.Close()
		wait.Done()
	}()
	go func() {
		io.Copy(conn, remote)
		conn.Close()
		wait.Done()
	}()
	wait.Wait()
	log.Printf("%s --> %s closed\n", conn.RemoteAddr(), s.remoteAddr)
}

func (s *sshSession) Run() error {
	listen, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConn(conn)
	}
}

func Login() (*ssh.Client, error) {
	var methods []ssh.AuthMethod
	if GlobalConfig().SSH.PrivateKey == "" && GlobalConfig().SSH.Password == "" {
		return nil, errors.New("empty private key and password")
	}
	if GlobalConfig().SSH.PrivateKey != "" {
		privateKeyContent, err := ioutil.ReadFile(GlobalConfig().SSH.PrivateKey)
		if err != nil {
			return nil, errors.New("unable to read private key")
		}
		signer, err := ssh.ParsePrivateKey(privateKeyContent)
		if err != nil {
			log.Fatalf("unable to parse private key:%v\n", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if GlobalConfig().SSH.Password != "" {
		methods = append(methods, ssh.Password(GlobalConfig().SSH.Password))
	}
	sshconfig := &ssh.ClientConfig{
		User:            GlobalConfig().SSH.User,
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", GlobalConfig().SSH.Adrr, sshconfig)
}

func SSHConnAndTransData() {
	sshClient, err := Login()
	if err != nil {
		log.Fatalf("init sshClient: %v\n", err)
	}

	for localPort, remote := range GlobalConfig().Ports {
		session := NewsshSession(":"+localPort, remote, sshClient)
		go func(local, remote string) {
			log.Printf("run session %s --> %s\n", local, remote)
			err := session.Run()
			if err != nil {
				log.Fatalf("run %s error:%s", local, err)
			}
		}(localPort, remote)
	}
}
