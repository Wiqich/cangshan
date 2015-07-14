package filepusher

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
	"golang.org/x/crypto/ssh"
)

func init() {
	application.RegisterModulePrototype("SSHPusher", new(SSHPusher))
}

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
		if strings.Index(s, ":") < 0 {
			s += ":22"
		}
		pusher.servers[i] = &server{
			sshConfig: &ssh.ClientConfig{
				User: pusher.Username,
				Auth: []ssh.AuthMethod{ssh.Password(pusher.Password)},
			},
			address: s,
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
				var client *ssh.Client
				logging.Debug("start dail to %s", s.address)
				client, err = ssh.Dial("tcp", s.address, s.sshConfig)
				if err != nil {
					logging.Debug("dail to %s fail: %s", s.address, err.Error())
					err = fmt.Errorf("new client to server %s fail: %s", s.address, err.Error())
					continue
				}
				logging.Debug("dail to %s success", s.address)
				defer client.Close()
				if err = scp(client, content, filename, remotePath+".tmp"); err != nil {
					err = fmt.Errorf("scp to %s fail: %s", remotePath+".tmp", err.Error())
					continue
				}
				if err = run(client, fmt.Sprintf("mv %s.tmp %s", remotePath, remotePath)); err != nil {
					err = fmt.Errorf("mv %s.tmp to %s fail: %s", remotePath, remotePath, err.Error())
					continue
				}
				logging.Debug("Push to server %s:%s success", s.address, remotePath)
				err = nil
				break
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

func scp(client *ssh.Client, content []byte, filename string, destination string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("New ssh session fail: %s", err.Error())
	}
	defer session.Close()
	var buf bytes.Buffer
	session.Stdin = &buf
	fmt.Fprintf(&buf, "C%#o %d %s\n", 0755, len(content), filename)
	buf.Write(content)
	buf.WriteByte('\x00')
	if err = session.Run("scp -t " + destination); err != nil {
		logging.Debug("scp -t %s fail: %s", destination, err.Error())
		return fmt.Errorf("run scp -t %s fail: %s", destination, err.Error())
	}
	logging.Debug("scp -t %s success", destination)
	return nil
}

func run(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("New ssh session fail: %s", err.Error())
	}
	defer session.Close()
	if err = session.Run(command); err != nil {
		logging.Debug("ssh %s")
		return fmt.Errorf("Run command %s fail: %s", command, err.Error())
	}
	return nil
}
