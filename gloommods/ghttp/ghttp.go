package ghttp

import (
	"../../gloommods/gbase"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Shopify/go-lua"
	"io/ioutil"
	"net"
	"net/http"
)

func luaDoTcp(l *lua.State) int {
	server, okServer := l.ToString(-3)
	port, okPort := l.ToString(-2)
	content, okContent := l.ToString(-1)

	if !okServer || !okPort || !okContent {
		l.PushNil()
		return 1
	}

	conn, err := net.Dial("tcp", server+":"+port)
	fmt.Printf("%s %s", server, port)

	if err != nil {
		l.PushString(err.Error())
		return 1
	}

	fmt.Fprintf(conn, content)
	reader := bufio.NewReader(conn)
	_, err = http.ReadResponse(reader, nil)

	if err != nil {
		l.PushString(err.Error())
		return 1
	}

	l.PushString("")
	return 1
}

func luaGetUrl(l *lua.State) int {
	url, ok := l.ToString(-1)
	if !ok {
		l.PushNil()
		return 1
	}
	resp, err := http.Get(url)

	if err != nil {
		l.PushString(err.Error())
		return 1
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		l.PushString(err.Error())
		return 1
	}

	body, _ = json.Marshal(resp)
	l.PushString(string(body))
	return 1
}

func registerLuaFunc(l *lua.State) {
	var funcs = []lua.RegistryFunction{
		{"getUrl", luaGetUrl},
		{"doTcp", luaDoTcp},
	}

	lua.NewLibrary(l, funcs)
	l.SetGlobal("ghttp")
}

func Load(server gbase.LuaProvider) {
	server.AddRegisterLuaFunc(registerLuaFunc)
}
