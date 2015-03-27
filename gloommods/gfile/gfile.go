package gfile

import (
	"../../gloommods/gbase"
	"github.com/Shopify/go-lua"
	"io/ioutil"
	"os"
)

func luaSaveFile(l *lua.State) int {
	name, okName := l.ToString(-2)
	content, okContent := l.ToString(-1)

	if !okName || !okContent {
		return 0
	}

	f, err := os.Create("./storage/" + name + ".dmp")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.Write([]byte(content))

	return 0
}

func luaAppendFile(l *lua.State) int {
	name, okName := l.ToString(-2)
	content, okContent := l.ToString(-1)

	if !okName || !okContent {
		return 0
	}

	f, err := os.OpenFile("./storage/"+name+".dmp",
		os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.Write([]byte(content))

	return 0
}

func luaLoadFile(l *lua.State) int {
	name, ok := l.ToString(-1)
	if !ok {
		return 0
	}
	f, err := os.Open("./storage/" + name + ".dmp")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	l.PushString(string(bytes))
	return 1
}

func registerLuaFunc(l *lua.State) {
	var funcs []lua.RegistryFunction

	funcSave := lua.RegistryFunction{
		Name:     "save",
		Function: luaSaveFile,
	}

	funcLoad := lua.RegistryFunction{
		Name:     "load",
		Function: luaLoadFile,
	}

	funcAppend := lua.RegistryFunction{
		Name:     "append",
		Function: luaAppendFile,
	}

	funcs = append(funcs, funcSave)
	funcs = append(funcs, funcLoad)
	funcs = append(funcs, funcAppend)

	lua.NewLibrary(l, funcs)
	l.SetGlobal("gfile")
}

func Load(server gbase.LuaProvider) {
	server.AddRegisterLuaFunc(registerLuaFunc)
}
