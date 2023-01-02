package jsapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	js "github.com/dop251/goja"
)

type FetchOptions struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

type FetchResponse struct {
	JSON func(call js.FunctionCall) js.Value `json:"json"`
	Text func(call js.FunctionCall) js.Value `json:"text"`
}

type Fetch struct {
	VM *js.Runtime
}

func (f *Fetch) Create() {
	f.VM.Set("fetch", f.fetchFn)
}

func (f *Fetch) fetchFn(call js.FunctionCall) js.Value {
	if len(call.Arguments) < 2 {
		return js.Undefined()
	}

	var options FetchOptions
	url := call.Argument(0).ToString().String()
	f.VM.ExportTo(call.Argument(1), &options)

	promise, resolve, reject := f.VM.NewPromise()
	fmt.Println(options.Method)

	go func() {
		if options.Method == "GET" {
			resp, err := http.Get(url)

			if err != nil {
				reject(err.Error())
				return
			}

			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			fetchResponse := FetchResponse{
				JSON: func(call js.FunctionCall) js.Value {
					promise, resolve, reject := f.VM.NewPromise()

					var obj interface{}
					err := json.Unmarshal(body, &obj)

					if err != nil {
						reject(err.Error())
					} else {
						resolve(obj)
					}

					return f.VM.ToValue(promise)
				},
				Text: func(call js.FunctionCall) js.Value {
					promise, resolve, _ := f.VM.NewPromise()
					resolve(string(body))
					return f.VM.ToValue(promise)
				},
			}

			resolve(fetchResponse)
		}
	}()

	return f.VM.ToValue(promise)
}
