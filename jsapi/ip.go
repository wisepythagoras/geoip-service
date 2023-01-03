package jsapi

import (
	"fmt"
	"net"

	js "github.com/dop251/goja"
)

type ipObj struct {
	IP net.IP
}

type JSIP struct {
	VM    *js.Runtime
	Proto *js.Object
}

func (ip *JSIP) Init() {
	cVal := ip.VM.ToValue(ip.constructor)
	ip.Proto = cVal.(*js.Object).Get("prototype").(*js.Object)

	ip.VM.Set("IP", cVal)
}

func (ip *JSIP) constructor(call js.ConstructorCall) *js.Object {
	if len(call.Arguments) == 0 {
		ip.VM.Interrupt(fmt.Errorf("an IP address is required"))
		return nil
	}

	ipAddress := call.Argument(0)
	inst := ip.VM.CreateObject(ip.Proto)
	fmt.Println(ipAddress)

	obj := ipObj{
		IP: net.ParseIP(ipAddress.String()),
	}

	inst.Set("ip", ipAddress)
	inst.Set("isValid", obj.IP != nil)

	// inst.Set("inc", func(call js.FunctionCall) js.Value {
	// 	obj.Val += 1
	// 	inst.Set("val", obj.Val)
	// 	return js.Undefined()
	// })

	return inst
}
