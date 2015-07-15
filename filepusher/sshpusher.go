package filepusher

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/coordination"
	"github.com/yangchenxing/cangshan/logging"
	"golang.org/x/crypto/ssh"
)

func init() {
	application.RegisterModulePrototype("SSHPusher", new(SSHPusher))
}

type server struct {
	sshConfig *ssh.ClientConfig
	address   string
}

type SSHPusher struct {
	sync.Mutex
	Coordination      coordination.Coordination
	ClusterName       string
	LocalPath         string
	RemotePath        string
	Username          string
	Password          string
	Retry             int
	WaitRetryInterval time.Duration
	servers           map[string]*ssh.ClientConfig
	clientConfig      *ssh.ClientConfig
}

func (pusher *SSHPusher) Initialize() error {
	if pusher.Coordination == nil {
		return errors.New("Missing Coordination")
	} else if pusher.ClusterName == "" {
		return errors.New("Missing ClusterName")
	}
	pusher.clientConfig = &ssh.ClientConfig{
		User: pusher.Username,
		Auth: []ssh.AuthMethod{ssh.Password(pusher.Password)},
	}
	pusher.servers = make(map[string]*ssh.ClientConfig)
	if nodes, err := pusher.Coordination.Discover(pusher.ClusterName); err != nil {
		return fmt.Errorf("Cannot discover cluster %s", pusher.ClusterName)
	} else {
		for _, node := range nodes {
			addr := node.Key
			if strings.Index(addr, ":") < 0 {
				addr += ":22"
			}
			pusher.servers[addr] = pusher.clientConfig
		}
	}
	if pusher.Retry == 0 {
		pusher.Retry = 1
	}
	go pusher.waitChange()
	return nil
}

func (pusher SSHPusher) Push(content []byte, localPath, remotePath string) error {
	pusher.Lock()
	defer pusher.Unlock()
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
	errChan := make(chan error, len(pusher.servers))
	for addr, sshConfig := range pusher.servers {
		go func() {
			var err error
			for i := 0; i < pusher.Retry; i++ {
				var client *ssh.Client
				logging.Debug("start dail to %s", addr)
				client, err = ssh.Dial("tcp", addr, sshConfig)
				if err != nil {
					logging.Debug("dail to %s fail: %s", addr, err.Error())
					err = fmt.Errorf("new client to server %s fail: %s", addr, err.Error())
					continue
				}
				logging.Debug("dail to %s success", addr)
				defer client.Close()
				if err = scp(client, content, filename, remotePath+".tmp"); err != nil {
					err = fmt.Errorf("scp to %s fail: %s", remotePath+".tmp", err.Error())
					continue
				}
				if err = run(client, fmt.Sprintf("mv %s.tmp %s", remotePath, remotePath)); err != nil {
					err = fmt.Errorf("mv %s.tmp to %s fail: %s", remotePath, remotePath, err.Error())
					continue
				}
				logging.Debug("Push to server %s:%s success", addr, remotePath)
				err = nil
				break
			}
			errChan <- err
		}()
	}
	var err error
	for range pusher.servers {
		if e := <-errChan; e != nil {
			err = e
			logging.Error("Push fail: %s", err.Error())
		}
	}
	return err
}

func (pusher *SSHPusher) waitChange() {
	receiveChan := make(chan *coordination.CoordinationEvent)
	stopChan := make(chan bool)
	errChan := make(chan error)
	go func() {
		errChan <- pusher.Coordination.LongWait(pusher.ClusterName, receiveChan, stopChan)
	}()
	for {
		for {
			select {
			case event := <-receiveChan:
				if event != nil {
					pusher.Lock()
					switch event.Type {
					case coordination.CreateNodeEvent:
						pusher.servers[event.Key] = pusher.clientConfig
					case coordination.DeleteNodeEvent:
						delete(pusher.servers, event.Key)
					}
					pusher.Unlock()
				}
			case err := <-errChan:
				logging.Error("Wait change receive error: %s", err.Error())
				break
			}
		}
		time.Sleep(pusher.WaitRetryInterval)
	}
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
