package jsapi

import (
	"fmt"
	"os"
	"path/filepath"

	js "github.com/dop251/goja"
)

type Storage struct {
	VM      *js.Runtime
	DataDir string
}

func (s *Storage) Init() {
	obj := s.VM.NewObject()
	obj.Set("init", s.initStorage)
	obj.Set("save", s.saveFile)
	obj.Set("remove", s.deleteFile)
	obj.Set("read", s.readFile)

	s.VM.Set("storage", obj)
}

func (s *Storage) initStorage(_ js.FunctionCall) js.Value {
	promise, resolve, reject := s.VM.NewPromise()
	_, err := os.Stat(s.DataDir)

	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(s.DataDir, 0700)

		if err != nil {
			reject(err.Error())
			return s.VM.ToValue(promise)
		}
	} else if err != nil {
		reject(err.Error())
		return s.VM.ToValue(promise)
	}

	resolve(js.Undefined())

	return s.VM.ToValue(promise)
}

func (s *Storage) saveFile(call js.FunctionCall) js.Value {
	promise, resolve, reject := s.VM.NewPromise()

	if len(call.Arguments) < 2 {
		err := fmt.Errorf("\"save\" requires 2 arguments, received %d", len(call.Arguments))
		reject(err.Error())
		return s.VM.ToValue(promise)
	}

	fileName := call.Argument(0).String()
	contents := call.Argument(1).String()

	file, err := os.Create(filepath.Join(s.DataDir, fileName))

	if err != nil {
		reject(err.Error())
		return s.VM.ToValue(promise)
	}

	_, err = file.Write([]byte(contents))

	if err != nil {
		reject(err.Error())
	} else {
		resolve(js.Undefined())
	}

	return s.VM.ToValue(promise)
}

func (s *Storage) deleteFile(call js.FunctionCall) js.Value {
	if len(call.Arguments) < 1 {
		return js.Undefined()
	}

	promise, resolve, reject := s.VM.NewPromise()
	fileName := call.Argument(0).String()

	if err := os.Remove(filepath.Join(s.DataDir, fileName)); err != nil {
		reject(err.Error())
	} else {
		resolve(js.Undefined())
	}

	return s.VM.ToValue(promise)
}

func (s *Storage) readFile(call js.FunctionCall) js.Value {
	if len(call.Arguments) < 1 {
		return js.Undefined()
	}

	promise, resolve, reject := s.VM.NewPromise()
	fileName := call.Argument(0).String()

	contents, err := os.ReadFile(filepath.Join(s.DataDir, fileName))

	if err != nil {
		reject(err.Error())
	} else {
		resolve(string(contents))
	}

	return s.VM.ToValue(promise)
}
