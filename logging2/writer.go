package logging

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type writer struct {
	sync.Once
	sync.Mutex
	Type     string
	Path     string
	Split    string
	file     *os.File
	Interval string
	interval time.Duration
	KeepTime string
	keepTime time.Duration
	EMail    struct {
		Server    string
		Sender    string
		Password  string
		Receivers []string
		Subject   string
		Delay     string
		delay     time.Duration
	}
	writer func(string)
}

func (writer *writer) initialize() (err error) {
	switch strings.ToLower(writer.Type) {
	case "timerotate":
		writer.writer, err = writer.newTimeRotate()
	case "email":
		writer.writer, err = writer.newEMail()
	case "stderr":
		writer.writer, err = writer.newStderr()
	default:
		err = fmt.Errorf("unknown writer type: %q", writer.Type)
	}
	return
}

func (writer *writer) write(text string) {
	if writer.writer != nil {
		writer.writer(text)
	}
}

func (writer *writer) newTimeRotate() (func(string), error) {
	var err error
	if writer.interval, err = time.ParseDuration(writer.Interval); err != nil {
		return nil, fmt.Errorf("invalid Interval: %q, %s", writer.Interval, err.Error())
	}
	if writer.keepTime, err = time.ParseDuration(writer.KeepTime); err != nil {
		return nil, fmt.Errorf("invalid KeepTime: %q, %s", writer.KeepTime, err.Error())
	}
	if info, err := os.Stat(writer.Path); err == nil {
		timestamp := info.ModTime().Truncate(writer.interval)
		if timestamp.Before(time.Now().Truncate(writer.interval)) {
			os.Rename(writer.Path, timestamp.Format(writer.Split))
		}
	}
	if writer.file, err = os.OpenFile(writer.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755); err != nil {
		return nil, err
	}
	writer.Once.Do(func() {
		go func() {
			for timestamp := time.Now().Truncate(writer.interval); ; timestamp = timestamp.Add(writer.interval) {
				time.Sleep(timestamp.Add(writer.interval).Sub(time.Now()))
				if err := writer.rotateTime(timestamp); err != nil {
					fmt.Fprintf(os.Stderr, "rotate log %q fail: %s\n", writer.Path, err.Error())
				}
				if err := writer.cleanTime(timestamp); err != nil {
					fmt.Fprintf(os.Stderr, "clean log %q fail: %s\n", writer.Path, err.Error())
				}
			}
		}()
	})
	return func(text string) {
		writer.Mutex.Lock()
		defer writer.Mutex.Unlock()
		if _, err := writer.file.WriteString(text); err != nil {
			fmt.Fprintf(os.Stderr, "write log %q fail: %s\n", writer.Path, err.Error())
		} else if err := writer.file.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "sync log %q fail: %s\n", writer.Path, err.Error())
		}
	}, nil
}

func (writer *writer) rotateTime(timestamp time.Time) error {
	var err error
	writer.Mutex.Lock()
	defer writer.Mutex.Unlock()
	writer.file.Sync()
	splitPath := timestamp.Format(writer.Split)
	if info, err := os.Stat(splitPath); err == nil && info != nil {
		return fmt.Errorf("log split file %q exists", splitPath)
	}
	if err := writer.file.Close(); err != nil {
		writer.file, _ = os.OpenFile(writer.Path, os.O_APPEND, 0755)
		return fmt.Errorf("close log file %q fail: %s", writer.Path, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "close log file %q success\n", writer.Path)
	}
	if err := os.Rename(writer.Path, splitPath); err != nil {
		writer.file, _ = os.OpenFile(writer.Path, os.O_APPEND, 0755)
		return fmt.Errorf("rename log file %q to %q fail: %s", writer.Path, splitPath, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "rename log file %q to %q success\n", writer.Path, splitPath)
	}
	if writer.file, err = os.OpenFile(writer.Path, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
		return fmt.Errorf("reopen log file %q fail: %s", writer.Path, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "reopen log file %q success\n", writer.Path)
	}
	return nil
}

func (writer *writer) cleanTime(timestamp time.Time) error {
	splitDir := filepath.Dir(writer.Split)
	infos, err := ioutil.ReadDir(splitDir)
	if err != nil {
		return fmt.Errorf("read directory %q fail: %s", splitDir, err.Error())
	}
	pattern := filepath.Base(writer.Split)
	for _, info := range infos {
		splitTime, err := time.Parse(pattern, info.Name())
		if err != nil {
			continue
		} else if timestamp.Sub(splitTime) <= writer.keepTime {
			continue
		}
		splitPath := filepath.Join(splitDir, info.Name())
		if err := os.Remove(splitPath); err != nil {
			return fmt.Errorf("remove log split file %q fail: %s", splitPath, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "remove log split file %q success\n", splitPath)
		}
	}
	return nil
}

func (writer *writer) newEMail() (func(string), error) {
	var err error
	if writer.EMail.delay, err = time.ParseDuration(writer.EMail.Delay); err != nil {
		return nil, fmt.Errorf("invalid email delay: %q, %s", writer.EMail.Delay, err.Error())
	}
	ch := make(chan string, 16)
	messages := list.New()
	go func() {
		for {
			message := <-ch
			messages.PushBack(message)
			ticker := time.NewTicker(writer.EMail.delay)
			for delay := true; delay; {
				select {
				case message := <-ch:
					messages.PushBack(message)
				case <-ticker.C:
					ticker.Stop()
					delay = false
					break
				}
			}
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n",
				writer.EMail.Sender,
				strings.Join(writer.EMail.Receivers, ","),
				writer.EMail.Subject)
			for elem := messages.Front(); elem != nil; elem = elem.Next() {
				buf.WriteString(elem.Value.(string))
			}
			auth := smtp.PlainAuth(writer.EMail.Sender, writer.EMail.Sender, writer.EMail.Password,
				strings.Split(writer.EMail.Server, ":")[0])
			if err := smtp.SendMail(writer.EMail.Server, auth, writer.EMail.Sender,
				writer.EMail.Receivers, buf.Bytes()); err != nil {
				fmt.Fprintf(os.Stderr, "send log email fail: %s\n", err.Error())
			} else {
				fmt.Fprintf(os.Stderr, "send log email success\n")
			}
			messages.Init()
		}
	}()
	return func(text string) {
		ch <- text
	}, nil
}

func (writer *writer) newStderr() (func(string), error) {
	return func(text string) {
		writer.Mutex.Lock()
		defer writer.Mutex.Unlock()
		os.Stderr.WriteString(text)
	}, nil
}
