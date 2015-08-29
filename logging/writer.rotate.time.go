package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("TimeRotateLogWriter", new(TimeRotateFileWriter))
}

type TimeRotateFileWriter struct {
	sync.Mutex
	sync.Once
	Path     string
	Split    string
	Interval time.Duration
	KeepTime time.Duration
	file     *os.File
}

func (w *TimeRotateFileWriter) Initialize() error {
	var err error
	if info, err := os.Stat(w.Path); err == nil {
		timestamp := info.ModTime().Truncate(w.Interval)
		if timestamp.Before(time.Now().Truncate(w.Interval)) {
			os.Rename(w.Path, timestamp.Format(w.Split))
		}
	}
	w.file, err = os.OpenFile(w.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	w.Once.Do(func() {
		go func() {
			for timestamp := time.Now().Truncate(w.Interval); ; timestamp = timestamp.Add(w.Interval) {
				time.Sleep(timestamp.Add(w.Interval).Sub(time.Now()))
				if err := w.rotate(timestamp); err != nil {
					Error("Rotate log %s fail: %s", w.Path, err.Error())
				}
				if err := w.clean(timestamp); err != nil {
					Error("Clean log %s fail: %s", w.Path, err.Error())
				}
			}
		}()
	})
	return nil
}

func (w *TimeRotateFileWriter) Write(b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	if n, err := w.file.Write(b); err != nil {
		return n, err
	}
	if err := w.file.Sync(); err != nil {
		return len(b), err
	}
	return len(b), nil
}

func (w *TimeRotateFileWriter) rotate(timestamp time.Time) error {
	var err error
	w.Lock()
	defer w.Unlock()
	w.file.Sync()
	splitPath := timestamp.Format(w.Split)
	if info, err := os.Stat(splitPath); err == nil && info != nil {
		fmt.Fprintf(os.Stderr, "log split file %s exist, cannot split", splitPath)
		return fmt.Errorf("log split file %s exist, cannot split", splitPath)
	}
	if err = w.file.Close(); err != nil {
		w.file, _ = os.OpenFile(w.Path, os.O_APPEND, 0755)
		fmt.Fprintf(os.Stderr, "Close log file %s fail: %s\n", w.Path, err.Error())
		return fmt.Errorf("Close log file %s fail: %s", w.Path, err.Error())
	} else {
		fmt.Fprintf(os.Stdout, "Close log file %s success\n", w.Path)
	}
	if err = os.Rename(w.Path, splitPath); err != nil {
		w.file, _ = os.OpenFile(w.Path, os.O_APPEND, 0755)
		fmt.Fprintf(os.Stderr, "Rename log file %s to %s fail: %s\n", w.Path, splitPath, err.Error())
		return fmt.Errorf("Rename log file %s fail: %s", w.Path, err.Error())
	} else {
		fmt.Fprintf(os.Stdout, "Rename log file %s to %s success\n", w.Path, splitPath)
	}
	if w.file, err = os.OpenFile(w.Path, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Reopen log fail %s fail: %s\n", w.Path, err.Error())
		return fmt.Errorf("Reopen log fail %s fail: %s", w.Path, err.Error())
	} else {
		fmt.Fprintf(os.Stdout, "Reopen log fail %s success\n", w.Path)
	}
	// fmt.Fprintf(os.Stderr, "Rotate log file success: %s -> %s\n", w.Path, splitPath)
	return nil
}

func (w *TimeRotateFileWriter) clean(timestamp time.Time) error {
	if w.KeepTime == 0 {
		return nil
	}
	splitDir := filepath.Dir(w.Split)
	infos, err := ioutil.ReadDir(splitDir)
	if err != nil {
		return fmt.Errorf("Read directory of %s fail: %s", w.Split, err.Error())
	}
	pattern := filepath.Base(w.Split)
	for _, info := range infos {
		splitTime, err := time.Parse(pattern, info.Name())
		if err != nil {
			continue
		}
		if timestamp.Sub(splitTime) <= w.KeepTime {
			continue
		}
		splitPath := filepath.Join(splitDir, info.Name())
		if err := os.Remove(splitPath); err != nil {
			fmt.Fprintf(os.Stderr, "Remove log file %s fail: %s\n", splitPath, err.Error())
			return fmt.Errorf("Remove log file %s fail: %s", splitPath, err.Error())
		} else {
			fmt.Fprintf(os.Stdout, "Remove log file %s success\n", splitPath)
		}
		// fmt.Fprintf(os.Stderr, "Remove log file %s success\n", splitPath)
	}
	return nil
}
