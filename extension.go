package main

import (
	"fmt"
	"os"
	"path/filepath"

	js "github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type EndpointReq struct {
	Param func(key string) string `json:"param"`
}

type EndpointRes struct {
	Send func(status int, resp ApiResponse) `json:"send"`
}

type EndpointDetails struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Handler  string `json:"handler"`
}

type ExtensionConfig struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Details any    `json:"details"`
}

type InstallFn func() ExtensionConfig
type HandlerFn func(req EndpointReq, res EndpointRes)

type Extension struct {
	extDir        string
	dir           os.DirEntry
	entry         os.DirEntry
	vm            *js.Runtime
	extType       string
	configDetails any
}

// Init will spin up the JS VM and run the script.
func (e *Extension) Init() error {
	e.vm = js.New()
	e.vm.SetFieldNameMapper(js.TagFieldNameMapper("json", true))

	entryPath := filepath.Join(e.extDir, e.dir.Name(), e.entry.Name())
	bytes, err := os.ReadFile(entryPath)

	if err != nil {
		return err
	}

	e.vm.Set("log", func(call js.FunctionCall) js.Value {
		for _, arg := range call.Arguments {
			fmt.Print(arg.String())
		}

		fmt.Println()

		return js.Undefined()
	})

	_, err = e.vm.RunScript(e.dir.Name(), string(bytes))

	if err != nil {
		return err
	}

	// The first step is to find and call the `install` function. This is a required part of
	// any extension and returns the extension's configuration.
	var installFn InstallFn
	err = e.vm.ExportTo(e.vm.Get("install"), &installFn)

	if err != nil {
		return err
	}

	res := installFn()
	e.extType = res.Type
	e.configDetails = res.Details

	return nil
}

// IsEndpointExtension returns true if this extension defines an endpoint.
func (e *Extension) IsEndpointExtension() bool {
	if len(e.extType) == 0 {
		return false
	}

	return e.extType == "endpoint"
}

// RegisterEndpoints will go through all of the endpoints and register them with gin.
func (e *Extension) RegisterEndpoints(r *gin.Engine) bool {
	if !e.IsEndpointExtension() {
		return false
	}

	var details []EndpointDetails
	mapstructure.Decode(e.configDetails, &details)

	for _, d := range details {
		if !e.registerEndpoint(r, d) {
			return false
		}
	}

	return true
}

func (e *Extension) registerEndpoint(r *gin.Engine, details EndpointDetails) bool {
	var handler HandlerFn
	err = e.vm.ExportTo(e.vm.Get(details.Handler), &handler)

	endpoint := filepath.Join("/api", details.Endpoint)

	endpointHandler := func(c *gin.Context) {
		req := EndpointReq{
			Param: c.Param,
		}
		res := EndpointRes{
			Send: func(status int, resp ApiResponse) {
				c.JSON(status, resp)
			},
		}

		handler(req, res)
	}

	if details.Method == "GET" {
		r.GET(endpoint, endpointHandler)
	} else if details.Method == "POST" {
		r.POST(endpoint, endpointHandler)
	} else if details.Method == "PUT" {
		r.PUT(endpoint, endpointHandler)
	} else if details.Method == "DELETE" {
		r.DELETE(endpoint, endpointHandler)
	}

	return true
}

func parseExtensions(path string) ([]*Extension, error) {
	files, err := os.ReadDir(path)
	extensions := []*Extension{}

	if err != nil {
		return extensions, err
	}

	for _, f := range files {
		if !f.IsDir() {
			break
		}

		extFiles, err := os.ReadDir(filepath.Join(path, f.Name()))

		if err != nil {
			return extensions, err
		}

		var entry os.DirEntry = nil

		for _, extFile := range extFiles {
			if extFile.IsDir() {
				continue
			}

			if extFile.Name() == "index.js" {
				entry = extFile
			}
		}

		if entry == nil {
			return extensions, fmt.Errorf("Extension folder %q doesn't have an entry point (index.js)", f.Name())
		}

		extension := &Extension{
			extDir: path,
			dir:    f,
			entry:  entry,
		}

		extensions = append(extensions, extension)
	}

	return extensions, nil
}
