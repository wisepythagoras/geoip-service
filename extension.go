package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	js "github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/wisepythagoras/geoip-service/jsapi"
)

type EndpointReq struct {
	Param     func(key string) string `json:"param"`
	GetHeader func(key string) string `json:"getHeader"`
	GetQuery  func(key string) string `json:"getQuery"`
}

type EndpointRes struct {
	JSON  func(status int, resp any)   `json:"json"`
	Abort func(status int, err string) `json:"abort"`
}

type EndpointDetails struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Handler  string `json:"handler"`
}

type CronJob struct {
	Cron string `json:"cron"`
	Job  string `json:"job"`
}

type ExtensionConfig struct {
	Version   int               `json:"version"`
	HasLookup bool              `json:"hasLookup"`
	Endpoints []EndpointDetails `json:"endpoints"`
	Jobs      []CronJob         `json:"jobs"`
	Name      string            `json:"name"`
}

type InstallFn func() ExtensionConfig
type HandlerFn func(req EndpointReq, res EndpointRes)

type Extension struct {
	extDir    string
	dir       os.DirEntry
	entry     os.DirEntry
	vm        *js.Runtime
	endpoints []EndpointDetails
	hasLookup bool
	scheduler *gocron.Scheduler
	name      string
	lookupFn  func(addr string) interface{}
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

	// Add all the APIs to the VM's runtime.

	consoleObj := jsapi.Console{VM: e.vm}
	consoleObj.Create()

	fetchFn := jsapi.Fetch{VM: e.vm}
	fetchFn.Create()

	ipListObj := jsapi.IPList{VM: e.vm}
	ipListObj.Init()

	ipObj := jsapi.JSIP{VM: e.vm}
	ipObj.Init()

	storageObj := jsapi.Storage{
		VM:      e.vm,
		DataDir: filepath.Join(e.extDir, e.dir.Name(), ".store"),
	}
	storageObj.Init()

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
	e.endpoints = res.Endpoints
	e.hasLookup = res.HasLookup
	e.name = res.Name

	if len(res.Name) == 0 {
		return fmt.Errorf("extension at %q doesn't have a name", e.dir.Name())
	}

	if e.hasLookup {
		err = e.vm.ExportTo(e.vm.Get("lookupIP"), &e.lookupFn)

		if err != nil {
			return err
		}
	}

	e.scheduler = gocron.NewScheduler(time.UTC)

	for _, job := range res.Jobs {
		var jobHandler func()
		err = e.vm.ExportTo(e.vm.Get(job.Job), &jobHandler)

		if err != nil {
			return err
		}

		fmt.Println("Registering", job.Cron, job.Job)
		e.scheduler.Cron(job.Cron).Do(jobHandler)
	}

	e.scheduler.StartAsync()

	return nil
}

// IsEndpointExtension returns true if this extension defines an endpoint.
func (e *Extension) IsEndpointExtension() bool {
	return len(e.endpoints) > 0
}

// IsLookupExtension returns true if this extension has lookup functions and should run
// on IP or domain lookup.
func (e *Extension) IsLookupExtension() bool {
	return e.hasLookup
}

// RunIPLookup will query the extension for data on a particular IP address.
func (e *Extension) RunIPLookup(ip string) (any, error) {
	if !e.IsLookupExtension() || e.lookupFn == nil {
		return nil, fmt.Errorf("this extension doesn't have lookup capabilities")
	}

	return e.lookupFn(ip), nil
}

// RegisterEndpoints will go through all of the endpoints and register them with gin.
func (e *Extension) RegisterEndpoints(r *gin.Engine) bool {
	if !e.IsEndpointExtension() {
		return false
	}

	for _, d := range e.endpoints {
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
		// We need a wait group because the JS VM may run an async handler and if we
		// don't wait here gin will exit the endpointHandler function and return a
		// 200 by default.
		wg := new(sync.WaitGroup)

		req := EndpointReq{
			Param:     c.Param,
			GetHeader: c.GetHeader,
			GetQuery:  c.Query,
		}
		res := EndpointRes{
			JSON: func(status int, resp any) {
				c.JSON(status, resp)
				wg.Done()
			},
			Abort: func(status int, err string) {
				c.AbortWithError(status, errors.New(err))
				wg.Done()
			},
		}

		wg.Add(1)
		handler(req, res)
		wg.Wait()
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
