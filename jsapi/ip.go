package jsapi

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	js "github.com/dop251/goja"
)

type IPObj struct {
	ParseList func(call js.FunctionCall) js.Value `json:"parseList"`
}

type JSIP struct {
	VM *js.Runtime
}

func (ip *JSIP) Init() {
	ipObj := IPObj{
		ParseList: ip.ParseList,
	}

	ip.VM.Set("IP", ipObj)
}

func (ip *JSIP) ParseList(call js.FunctionCall) js.Value {
	if len(call.Arguments) < 1 {
		return js.Undefined()
	}

	promise, resolve, reject := ip.VM.NewPromise()
	strList := call.Argument(0).ToString().String()

	// Parses the list into a string array so that the JS engine will be able to interpret that.
	list, err := parseSimpleIPList(strList)

	if err != nil {
		reject(err.Error())
	} else {
		resolve(list)
	}

	return ip.VM.ToValue(promise)
}

func parseSimpleIPList(list string) ([]string, error) {
	lines := strings.Split(list, "\n")
	ipAddresses := []string{}

	for _, line := range lines {
		re := regexp.MustCompile(`(#(?:[^\n]+)?)`)
		line = strings.Trim(re.ReplaceAllLiteralString(line, ""), " ")

		if len(line) == 0 {
			continue
		}

		re = regexp.MustCompile(`(/\d+)`)
		ip := net.ParseIP(re.ReplaceAllLiteralString(line, ""))

		if ip == nil {
			return nil, fmt.Errorf("invalid IP address %q", line)
		}

		ipAddresses = append(ipAddresses, line)
	}

	return ipAddresses, nil
}