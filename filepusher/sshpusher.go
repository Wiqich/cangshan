package filepusher

import (
	"bytes"
	"fmt"
	"github.com/tmc/scp"
	"github.com/yangchenxing/cangshan/logging"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

var (
	serverPattern = regexp.MustCompile("((?P<username>[^:]*):(?P<password>[^@])@)?(?P<address>.*)")
)

type server struct {
	sshConfig *ssh.ClientConfig
	address   string
}

type SSHPusher struct {
	Servers    []string
	LocalPath  string
	RemotePath string
	Username   string
	Password   string
	Retry      int
	servers    []*server
}

func (pusher *SSHPusher) Initialize() error {
	pusher.servers = make([]*server, len(pusher.Servers))
	for i, s := range pusher.Servers {
		submatch := serverPattern.FindStringSubmatch(s)
		if submatch == nil {
			return fmt.Errorf("Invalid server: %s", s)
		}
		username := pusher.Username
		password := pusher.Password
		var address string
		for i, name := range serverPattern.SubexpNames() {
			switch name {
			case "username":
				username = submatch[i]
			case "password":
				password = submatch[i]
			case "address":
				address = submatch[i]
			}
		}
		pusher.servers[i] = &server{
			sshConfig: &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{ssh.Password(password)},
			},
			address: address,
		}
	}
	if pusher.Retry == 0 {
		pusher.Retry = 1
	}
	return nil
}

func (pusher SSHPusher) Push(content []byte, localPath, remotePath string) error {
	if localPath != "" {
		localPath = filepath.Join(pusher.LocalPath, localPath)
		if err := ioutil.WriteFile(localPath, content, 0755); err != nil {
			logging.Error("Save local file %s fail: %s", localPath, err.Error())
		}
	}
	if remotePath != "" {
		remotePath = filepath.Join(pusher.RemotePath, remotePath)
	} else {
		remotePath = pusher.RemotePath
	}
	filename := filepath.Base(remotePath)
	errChan := make(chan error, len(pusher.Servers))
	for _, s := range pusher.servers {
		go func() {
			var err error
			for i := 0; i < pusher.Retry; i++ {
				if client, err := ssh.Dial("tcp", s.address, s.sshConfig); err != nil {
					err = fmt.Errorf("new client to server %s fail: %s", s.address, err.Error())
				} else if session, err := client.NewSession(); err != nil {
					err = fmt.Errorf("new session to server %s fail: %s", s.address, err.Error())
				} else if err := scp.Copy(int64(len(content)), 0755, filename, bytes.NewReader(content), remotePath+".tmp", session); err != nil {
					err = fmt.Errorf("scp to server %s fail: %s", s.address, err.Error())
				} else if err := session.Run(fmt.Sprintf("mv %s.tmp %s", remotePath, remotePath)); err != nil {
					err = fmt.Errorf("ssh run on server %s fail: %s", s.address, err.Error())
				}
				logging.Debug("Push to server %s:%s success", s.address, remotePath)
				errChan <- nil
				return
			}
			errChan <- err
		}()
	}
	var err error
	for range pusher.Servers {
		if e := <-errChan; e != nil {
			err = e
			logging.Error("Push fail: %s", err.Error())
		}
	}
	return err
}
