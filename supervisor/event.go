package supervisor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	abstractTypes = []string{
		"PROCESS_COMMUNICATION",
		"PROCESS_LOG",
		"PROCESS_STATE",
		"SUPERVISOR_STATE_CHANGE",
		"TICK",
	}
)

type Event map[string]string

func NewEvent(reader io.Reader) (Event, error) {
	var event Event = make(map[string]string)
	buf := bufio.NewReader(reader)
	if data, err := buf.ReadBytes('\n'); err != nil {
		return nil, fmt.Errorf("read header fail: %s", err.Error())
	} else {
		event.parseMap(data)
	}
	payload := make([]byte, event.Len())
	if _, err := buf.Read(payload); err != nil {
		return nil, fmt.Errorf("read body fail: %s", err.Error())
	} else if index := bytes.IndexByte(payload, '\n'); index >= 0 {
		event.parseMap(payload[:index])
		event["payload"] = string(payload[index+1:])
	} else {
		event["payload"] = string(payload)
	}
	return event, nil
}

func (event Event) String() string {
	var buf bytes.Buffer
	buf.WriteString("Event{")
	first := true
	for key, value := range event {
		if key == "payload" {
			continue
		}
		if first {
			first = false
		} else {
			buf.WriteRune(' ')
		}
		buf.WriteString(key)
		buf.WriteRune(':')
		buf.WriteString(value)
	}
	buf.WriteRune('}')
	return buf.String()
}

func (event Event) Format(s fmt.State, r rune) {
	s.Write([]byte(event.String()))
}

func (event Event) AbstractType() string {
	name := event.Name()
	for _, abstractType := range abstractTypes {
		if strings.HasPrefix(name, abstractType) {
			return abstractType
		}
	}
	return name
}

func (event Event) State() string {
	switch typ := event.AbstractType(); typ {
	case "PROCESS_TYPE":
		fallthrough
	case "SUPERVISOR_STATE_CHANGE":
		return event.Name()[len(typ)+1:]
	}
	return ""
}

func (event Event) Name() string {
	return event["eventname"]
}

func (event Event) Serial() int64 {
	return event.getInt("serial")
}

func (event Event) Pool() string {
	return event["pool"]
}

func (event Event) Version() string {
	return event["version"]
}

func (event Event) PoolSerial() int64 {
	return event.getInt("poolserial")
}

func (event Event) Len() int64 {
	return event.getInt("len")
}

func (event Event) parseMap(data []byte) {
	for _, token := range strings.Split(strings.TrimSpace(string(data)), " ") {
		if token == "" {
			continue
		}
		switch pair := strings.SplitN(token, ":", 2); len(pair) {
		case 2:
			event[pair[0]] = pair[1]
		case 1:
			event[pair[0]] = ""
		}
	}
}

func (event Event) getInt(key string) int64 {
	if value, found := event[key]; found {
		if value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return value
		}
	}
	return 0
}
