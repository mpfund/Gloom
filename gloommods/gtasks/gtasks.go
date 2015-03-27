package gtasks

import (
	"../../gloommods/gbase"
	"bytes"
	"container/list"
	"crypto/rand"
	"fmt"
	"github.com/Shopify/go-lua"
	"os/exec"
	"time"
)

type LongRunningTask struct {
	Id        string
	Name      string
	Cmd       string
	Args      []string
	Out       string
	ExecTime  int
	StartTime time.Time
	IsLua     bool
}

func NewTask() *LongRunningTask {
	t := new(LongRunningTask)
	t.Id = CreateId()
	t.StartTime = time.Now()
	return t
}

func runTaskLoop(listQueue *list.List, doneQueue *list.List, server gbase.LuaProvider) {
	for {
		time.Sleep(1 * time.Second)
		element := listQueue.Front()
		if element != nil {
			task := element.Value.(*LongRunningTask)
			listQueue.Remove(element)

			if task.StartTime.Sub(time.Now()).Seconds() > 0 {
				continue
			}

			if !task.IsLua {
				cmd := exec.Command(task.Cmd, task.Args...)
				var out bytes.Buffer
				cmd.Stdout = &out
				cmd.Start()
				cmd.Wait()
				task.Out = out.String()
			} else {
				l, err := server.GetLuaState()
				if err != nil {
					panic(err)
				}

				lua.DoString(l, task.Cmd)
				ret, ok := l.ToString(-1)
				if ok {
					task.Out = ret
					l.Pop(1)
				}
			}

			task.ExecTime = 0
			doneQueue.PushFront(task)
			fmt.Printf("task %s", task.Out)
		}
	}
}

func runTasks(listQueue *list.List, doneQueue *list.List, server gbase.LuaProvider) {
	go runTaskLoop(listQueue, doneQueue, server)
}

func CreateId() string {
	size := 32 // change the length of the generated random string here

	rb := make([]byte, size)
	rand.Read(rb)

	var dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	for k, v := range rb {
		rb[k] = dictionary[v%byte(len(dictionary))]
	}

	//rs := base64.URLEncoding.EncodeToString(rb)
	return string(rb)
}

func luaQueueCommand(q *list.List) func(*lua.State) int {
	taskQueue := q

	return func(l *lua.State) int {
		cmd, ok := l.ToString(-1)
		if !ok {
			return 0
		}
		task := NewTask()
		task.Name = cmd
		task.Cmd = cmd

		taskQueue.PushBack(task)
		l.PushString(task.Id)
		return 1
	}
}

func luaQueueLuaCommand(q *list.List) func(*lua.State) int {
	taskQueue := q

	return func(l *lua.State) int {
		cmd, ok := l.ToString(-1)
		if !ok {
			return 0
		}

		task := NewTask()
		task.Name = cmd
		task.Cmd = cmd
		task.IsLua = true

		taskQueue.PushBack(task)
		l.PushString(task.Id)
		return 1
	}
}

func luaGetCommandOutputId(q *list.List) func(*lua.State) int {
	doneQueue := q

	return func(l *lua.State) int {
		id, ok := l.ToString(-1)
		if !ok {
			l.PushNil()
			return 1
		}

		for e := doneQueue.Front(); e != nil; e = e.Next() {
			task := e.Value.(*LongRunningTask)
			if id == task.Id {
				l.PushString(task.Out)
				return 1
			}
		}

		l.PushNil()
		return 0
	}

}

func Load(server gbase.LuaProvider) {
	// uses lists because we want to a rearrange, pause and continue
	// functionality otherwise we could use channels
	taskQueue := list.New()
	doneQueue := list.New()
	runTasks(taskQueue, doneQueue, server)

	server.AddRegisterLuaFunc(func(l *lua.State) {
		var funcs = []lua.RegistryFunction{
			{"queueCommand", luaQueueCommand(taskQueue)},
			{"queueLuaCommand", luaQueueLuaCommand(taskQueue)},
			{"getCommandById", luaGetCommandOutputId(doneQueue)},
		}

		lua.NewLibrary(l, funcs)
		l.SetGlobal("gtasks")
	})

}
