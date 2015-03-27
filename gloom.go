// gloom
package main

import (
	"./gloommods/gfile"
	"./gloommods/ghttp"
	"./gloommods/gtasks"
	"fmt"
	"github.com/Shopify/go-lua"
	"io/ioutil"
	"net/http"
)

type GloomServer struct {
	luaFuncs []func(*lua.State)
}

func (s *GloomServer) GetLuaFuncs() []func(*lua.State) {
	return s.luaFuncs
}

func (s *GloomServer) AddRegisterLuaFunc(fc func(*lua.State)) {
	s.luaFuncs = append(s.luaFuncs, fc)
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s GloomServer) GetLuaState() (*lua.State, error) {
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

	funcs := s.GetLuaFuncs()
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
	if !ok {
		return "", false
	}
	l.Pop(1)
	return ret, true
}

func LuaHandler(server *GloomServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		l, err := server.GetLuaState()
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		ret, _ := CallHandleRequest(l, req)

		fmt.Fprintf(w, ret)
	}
}

func main() {
	server := GloomServer{}

	gtasks.Load(&server)
	ghttp.Load(&server)
	gfile.Load(&server)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static/")))
	mux.HandleFunc("/d/", LuaHandler(&server))
	http.ListenAndServe(":82", mux)
}
