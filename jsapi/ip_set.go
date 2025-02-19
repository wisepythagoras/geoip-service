package jsapi

import (
	"fmt"

	js "github.com/dop251/goja"
)

type JSIPSet struct {
	VM    *js.Runtime
	Proto *js.Object
}

func (ip *JSIPSet) Init() {
	cVal := ip.VM.ToValue(ip.constructor)
	ip.Proto = cVal.(*js.Object).Get("prototype").(*js.Object)

	ip.VM.Set("IPSet", cVal)
}

func (ip *JSIPSet) constructor(call js.ConstructorCall) *js.Object {
	if len(call.Arguments) == 0 {
		ip.VM.Interrupt(fmt.Errorf("an IPSet is required"))
		return nil
	}

	var err error
	ipAddrsInterface := call.Argument(0).Export()
	ipAddrs, ok := ipAddrsInterface.([]string)

	if !ok {
		ipAddrs, err = parseSimpleIPList(call.Argument(0).String())

		if err != nil {
			ip.VM.Interrupt(err)
			return nil
		}
	}

	fmt.Println(ipAddrs)

	inst := ip.VM.CreateObject(ip.Proto)

	return inst
}
