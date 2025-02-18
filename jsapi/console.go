package jsapi

import (
	"fmt"

	js "github.com/dop251/goja"
	"github.com/fatih/color"
)

const (
	PRINT_NORMAL = 0
	PRINT_ERROR  = 1
	PRINT_WARN   = 2
)

type consoleObj struct {
	Log   func(call js.FunctionCall) js.Value `json:"log"`
	Warn  func(call js.FunctionCall) js.Value `json:"warn"`
	Error func(call js.FunctionCall) js.Value `json:"error"`
}

type Console struct {
	VM *js.Runtime
}

func (c *Console) Create() {
	console := consoleObj{
		Log:   c.LogFactory(PRINT_NORMAL),
		Warn:  c.LogFactory(PRINT_WARN),
		Error: c.LogFactory(PRINT_ERROR),
	}

	c.VM.Set("console", console)
}

func (c *Console) LogFactory(level uint) func(call js.FunctionCall) js.Value {
	return func(call js.FunctionCall) js.Value {
		for _, arg := range call.Arguments {
			if level == PRINT_ERROR {
				c := color.New(color.FgRed)
				c.Print(arg, " ")
			} else if level == PRINT_WARN {
				c := color.New(color.FgYellow)
				c.Print(arg, " ")
			} else {
				fmt.Print(arg, " ")
			}
		}

		fmt.Println()

		return js.Undefined()
	}
}
