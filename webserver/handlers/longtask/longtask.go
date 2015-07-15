package longtask

import (
	"bytes"
	"container/list"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterModulePrototype("WebServerLongTaskExecutive", new(LongTaskExecutive))
	application.RegisterBuiltinModule("WebServerLongTaskStatus", new(LongTaskStatus))
}

var (
	CleanInterval = time.Hour
	MaxAge        = time.Hour * 24
	tasks         = make(map[int]*Task)
	once          sync.Once
	nextTaskID    = 1
)

type TaskStatus int

const (
	Ready TaskStatus = iota
	Running
	Done
)

type Command interface {
	Run(request *webserver.Request, taskID int, output io.Writer) error
}

type Task struct {
	ID        int
	Request   *webserver.Request
	Command   Command
	Status    TaskStatus
	Output    bytes.Buffer
	Error     error
	BeginTime time.Time
	EndTime   time.Time
}

func (task *Task) Run() {
	if task.Status != Ready {
		return
	}
	task.BeginTime = time.Now()
	task.Error = task.Command.Run(task.Request, task.ID, &task.Output)
	task.EndTime = time.Now()
}

type LongTaskExecutive struct {
	sync.Mutex
	Command Command
}

func (ex *LongTaskExecutive) Handle(request *webserver.Request) {
	once.Do(func() {
		go clean()
	})
	ex.Lock()
	defer ex.Unlock()
	task := &Task{
		ID:      nextTaskID,
		Request: request,
		Command: ex.Command,
		Status:  Ready,
	}
	nextTaskID++
	go task.Run()
	webserver.WriteStandardJSONResult(request, true, "entities", map[string]interface{}{"id": task.ID})
}

type LongTaskStatus struct{}

func (handler LongTaskStatus) Handle(request *webserver.Request) {
	outputOffset, _ := strconv.Atoi(request.Param["offset"].(string))
	if id, ok := request.Param["id"].(string); !ok {
		webserver.WriteStandardJSONResult(request, false, "message", "missing task id")
	} else if id, err := strconv.Atoi(id); err != nil {
		webserver.WriteStandardJSONResult(request, false, "message", "invalid task id")
	} else if task := tasks[id]; task == nil {
		webserver.WriteStandardJSONResult(request, false, "message", "unknown task id")
	} else {
		entity := map[string]interface{}{
			"id":        task.ID,
			"done":      task.Status == Done,
			"beginTime": task.BeginTime.Format("2006-01-02:15:04:05"),
			"output":    task.Output.Bytes()[outputOffset:],
		}
		if task.Status == Done {
			entity["entTime"] = task.EndTime.Format("2006-01-02:15:04:05")
			entity["success"] = task.Error == nil
			if task.Error != nil {
				entity["error"] = task.Error.Error()
			}
		}
		webserver.WriteStandardJSONResult(request, true, "entities", entity)
	}
}

func clean() {
	for {
		time.Sleep(CleanInterval)
		expiredIds := list.New()
		now := time.Now()
		for id, task := range tasks {
			if task.BeginTime.Add(MaxAge).Before(now) {
				expiredIds.PushBack(id)
			}
		}
		for id := expiredIds.Front(); id != nil; id = id.Next() {
			delete(tasks, id.Value.(int))
		}
	}
}
