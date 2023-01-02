package jsapi

import (
	"fmt"

	js "github.com/dop251/goja"
)

type consoleObj struct {
	Log   func(call js.FunctionCall) js.Value `json:"log"`
	Error func(call js.FunctionCall) js.Value `json:"error"`
}

type Console struct {
	VM *js.Runtime
}

func (c *Console) Create() {
	console := consoleObj{
		Log:   c.Log,
		Error: c.Log,
	}

	c.VM.Set("console", console)
}

func (c *Console) Log(call js.FunctionCall) js.Value {
	for _, arg := range call.Arguments {
		fmt.Print(arg.String())
	}

	fmt.Println()

	return js.Undefined()
}
