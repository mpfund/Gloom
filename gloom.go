// gloom
package main

import (
	"bytes"
	"container/list"
	"crypto/rand"
	"fmt"
	"github.com/Shopify/go-lua"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

type LongRunningTask struct {
	Id       string
	Name     string
	Cmd      string
	Args     []string
	Out      string
	ExecTime int
}

func NewTask() *LongRunningTask {
	t := new(LongRunningTask)
	t.Id = CreateId()
	return t
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

func runTask(listQueue *list.List, doneQueue *list.List) {
	go func() {
		for {
			element := listQueue.Front()
			if element != nil {
				task := element.Value.(*LongRunningTask)
				listQueue.Remove(element)
				cmd := exec.Command(task.Cmd, task.Args...)
				var out bytes.Buffer
				cmd.Stdout = &out
				cmd.Start()
				cmd.Wait()
				task.Out = out.String()
				task.ExecTime = 0
				doneQueue.PushFront(task)
				fmt.Printf("task %s", task.Out)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func getLua() (*lua.State, error) {
	l := lua.NewState()
	lua.OpenLibraries(l)
	files, err := ioutil.ReadDir("./lua/")
	checkerr(err)

	for f := range files {
		err := lua.DoFile(l, "./lua/"+files[f].Name())
		if err != nil {
			return nil, err
		}
	}

	funcs := GetLuaFuncs()
	for f := range funcs {
		reg := funcs[f]
		if reg != nil {
			reg(l)
		}
	}

	return l, nil
}

func CallHandleRequest(l *lua.State, req *http.Request) (string, bool) {
	l.Global("HandleRequest")
	exists := l.IsFunction(-1)
	if !exists {
		l.Remove(-1)
		return "", false
	}

	l.PushString(req.URL.String())
	l.Call(1, 1)
	ret, ok := l.ToString(-1)
	l.Pop(1)
	if !ok {
		return "", false
	}
	return ret, true
}

func LuaHandler(w http.ResponseWriter, req *http.Request) {
	l, err := getLua()
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	ret, _ := CallHandleRequest(l, req)
	fmt.Fprintf(w, ret)
}

type RegisterLuaFunc func(*lua.State)

var luaFuncs []RegisterLuaFunc

func GetLuaFuncs() []RegisterLuaFunc {
	return luaFuncs
}

func AddRegisterLuaFunc(fc RegisterLuaFunc) {
	if luaFuncs == nil {
		luaFuncs = make([]RegisterLuaFunc, 10)
	}
	luaFuncs = append(luaFuncs, fc)
}

func RegisterTaskHandler() {
	// lists because we want to rearrange, pause and continue
	// functionality otherwise we could use channels
	taskQueue := list.New()
	doneQueue := list.New()
	runTask(taskQueue, doneQueue)

	AddRegisterLuaFunc(func(l *lua.State) {
		lua.NewLibrary(l, []lua.RegistryFunction{{
			Name: "queueCommand",
			Function: func(l *lua.State) int {
				cmd, ok := l.ToString(-1)
				if !ok {
					return 0
				}
				task := NewTask()
				task.Name = cmd
				task.Cmd = cmd

				taskQueue.PushBack(task)
				fmt.Printf("Queued command %v", task)
				l.PushString(task.Id)
				return 1
			}},
			{
				Name: "getCommandOutputId",
				Function: func(l *lua.State) int {
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
				}},
		})

		l.SetGlobal("server")
	})
}

func main() {
	RegisterTaskHandler()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static/")))
	mux.HandleFunc("/d/", LuaHandler)
	http.ListenAndServe(":82", mux)
}
