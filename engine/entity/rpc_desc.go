package entity

import (
	"reflect"
	"strings"
)

const (
	rfServer = 1 << iota
	rfOwnClient
	rfOtherClient
)

type rpcDesc struct {
	Func       reflect.Value
	Flags      uint
	MethodType reflect.Type
	NumArgs    int
}

type rpcDescMap map[string]*rpcDesc

func (rdm rpcDescMap) visit(method reflect.Method) {
	methodName := method.Name
	var flag uint
	var rpcName string
	if strings.HasSuffix(methodName, "_Client") {
		flag |= rfServer + rfOwnClient
		rpcName = methodName[:len(methodName)-7]
	} else if strings.HasSuffix(methodName, "_AllClients") {
		flag |= rfServer + rfOwnClient + rfOtherClient
		rpcName = methodName[:len(methodName)-11]
	} else {
		// server method
		flag |= rfServer
		rpcName = methodName
	}

	methodType := method.Type
	rdm[rpcName] = &rpcDesc{
		Func:       method.Func,
		Flags:      flag,
		MethodType: methodType,
		NumArgs:    methodType.NumIn() - 1, // do not count the receiver
	}
}
