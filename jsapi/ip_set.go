package jsapi

import (
	"fmt"
	"net"
	"strings"
	"time"

	js "github.com/dop251/goja"
)

const (
	UPDATE_HOURLY   = 0
	UPDATE_DAILY    = 1
	UPDATE_WEEKLY   = 2
	UPDATE_BIWEEKLY = 3
	UPDATE_MONTHLY  = 4
)

type ipSetOpts struct {
	Name        string    `json:"name"`
	Maintainer  string    `json:"maintainer"`
	URL         string    `json:"url"`
	Date        time.Time `json:"date"`
	UpdateFreq  uint      `json:"update_freq"`
	Version     uint64    `json:"version"`
	Description string    `json:"description"`
	Notes       string    `json:"notes"`
}

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
	var opts ipSetOpts

	entries := make([]string, 0)
	err = ip.VM.ExportTo(call.Argument(0), &entries)
	ipAddrsMap := make(map[string]net.IP)
	cidrEntries := make([]*net.IPNet, 0)
	cidrEntryMap := make(map[string]*net.IPNet)

	if len(call.Arguments) > 1 {
		ip.VM.ExportTo(call.Argument(1), &opts)
	}

	if err != nil {
		entries, err = parseSimpleIPList(call.Argument(0).String())

		if err != nil {
			ip.VM.Interrupt(err)
			return nil
		}
	}

	for _, entry := range entries {
		if strings.Contains(entry, "/") {
			_, cidrEntry, err := net.ParseCIDR(entry)

			if err != nil {
				ip.VM.Interrupt(err)
				return nil
			}

			cidrEntries = append(cidrEntries, cidrEntry)
			cidrEntryMap[cidrEntry.String()] = cidrEntry
		} else {
			ipAddrsMap[entry] = net.ParseIP(entry)
		}
	}

	inst := ip.VM.CreateObject(ip.Proto)

	inst.Set("getEntries", func(_ js.FunctionCall) js.Value {
		return ip.VM.ToValue(entries)
	})

	inst.Set("contains", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			return ip.VM.ToValue(false)
		}

		entry := call.Argument(0).String()

		if !strings.Contains(entry, "/") {
			ipEntry := net.ParseIP(entry)

			if _, ok := ipAddrsMap[entry]; ok {
				return ip.VM.ToValue(true)
			}

			for _, cidrEntry := range cidrEntries {
				if cidrEntry.Contains(ipEntry) {
					return ip.VM.ToValue(true)
				}
			}

			return ip.VM.ToValue(false)
		}

		_, cidrEntry, err := net.ParseCIDR(entry)

		if err != nil {
			ip.VM.Interrupt(err)
			return ip.VM.ToValue(false)
		}

		if _, ok := cidrEntryMap[cidrEntry.String()]; ok {
			return ip.VM.ToValue(true)
		}

		return ip.VM.ToValue(false)
	})

	// inst.Set("append", func(call js.FunctionCall) js.Value {

	// })

	return inst
}
