package jsapi

import (
	"fmt"
	"net"
	"strings"

	js "github.com/dop251/goja"
)

type ipObj struct {
	IP      net.IP
	IPRange *net.IPNet
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

	ipAddress := call.Argument(0).String()
	isCidrRange := strings.Contains(ipAddress, "/")

	var netIP net.IP
	var ipRange *net.IPNet
	var err error

	if isCidrRange {
		netIP, ipRange, err = net.ParseCIDR(ipAddress)
	} else {
		netIP = net.ParseIP(ipAddress)
	}

	if err != nil {
		ip.VM.Interrupt(err)
		return nil
	}

	obj := ipObj{
		IP:      netIP,
		IPRange: ipRange,
	}

	inst := ip.VM.CreateObject(ip.Proto)
	inst.Set("ip", ipAddress)
	inst.Set("isValid", obj.IP != nil)
	inst.Set("isCIDRRange", isCidrRange && ipRange != nil)
	inst.Set("isLoopback", func(_ js.FunctionCall) js.Value {
		return ip.VM.ToValue(obj.IP.IsLoopback())
	})
	inst.Set("isPrivate", func(_ js.FunctionCall) js.Value {
		return ip.VM.ToValue(obj.IP.IsPrivate())
	})
	inst.Set("isUnspecified", func(_ js.FunctionCall) js.Value {
		return ip.VM.ToValue(obj.IP.IsUnspecified())
	})
	inst.Set("getMask", func(_ js.FunctionCall) js.Value {
		return ip.VM.ToValue(obj.IP.DefaultMask().String())
	})
	inst.Set("contains", func(call js.FunctionCall) js.Value {
		if !isCidrRange || len(call.Arguments) < 1 {
			return ip.VM.ToValue(false)
		}

		ipAddress := call.Argument(0).String()
		netIP := net.ParseIP(ipAddress)

		return ip.VM.ToValue(ipRange.Contains(netIP))
	})

	// inst.Set("inc", func(call js.FunctionCall) js.Value {
	// 	obj.Val += 1
	// 	inst.Set("val", obj.Val)
	// 	return js.Undefined()
	// })

	return inst
}
