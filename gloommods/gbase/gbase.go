package gbase

import (
	"github.com/Shopify/go-lua"
)

type LuaProvider interface {
	GetLuaState() (*lua.State, error)
	AddRegisterLuaFunc(func(*lua.State))
}
