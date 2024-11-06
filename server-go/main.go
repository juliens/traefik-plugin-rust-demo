package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/stealthrocket/wasi-go/imports"
	wazergo_wasip1 "github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wazergo"
	wazeroapi "github.com/tetratelabs/wazero/api"

	"github.com/http-wasm/http-wasm-host-go/api"
	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"github.com/tetratelabs/wazero"
)

type logger struct {
}

func (l logger) IsEnabled(level api.LogLevel) bool {
	return true
}

func (l logger) Log(ctx context.Context, level api.LogLevel, s string) {
	fmt.Printf("[%s] %s\n", level, s)
}

func main() {
	go func() {
		monitorMemory(5*time.Second, nil)
	}()

	code, err := os.ReadFile("./plugin.wasm")
	if err != nil {
		log.Fatal(err)
	}

	conf := []byte(`{"headerName":"test", "headers":{
	"X-Field":"Value"
}}`)

	cache, err := wazero.NewCompilationCacheWithDir("/tmp/cachewasm")
	if err != nil {
		log.Fatal(err)
	}
	var applyCtx func(context.Context) context.Context
	applyCtx = func(ctx context.Context) context.Context {
		return ctx
	}
	ctx := context.Background()

	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(cache).WithDebugInfoEnabled(false))

	guestModule, err := rt.CompileModule(ctx, code)
	if err != nil {
		log.Fatal(fmt.Errorf("Error while compile guestModule: %w", err))
	}

	if extension := imports.DetectSocketsExtension(guestModule); extension != nil {
		fmt.Println("extension")
		builder := imports.NewBuilder().WithSocketsExtension("auto", guestModule)
		ctx, sys, err := builder.Instantiate(ctx, rt)
		if err != nil {
			log.Fatal(fmt.Errorf("Instantiate builder: %w", err))
		}

		// builder.WithDirs("/:/")

		inst, err := wazergo.Instantiate(ctx, rt, wazergo_wasip1.NewHostModule(*extension), wazergo_wasip1.WithWASI(sys))
		if err != nil {
			log.Fatal(fmt.Errorf("wazergo instantiation: %w", err))
		}
		applyCtx = func(ctx context.Context) context.Context {
			return wazergo.WithModuleInstance(ctx, inst)
		}
	}

	mw, err := wasm.NewMiddleware(applyCtx(ctx), code, handler.Logger(logger{}),
		handler.ModuleConfig(wazero.NewModuleConfig().WithStartFunctions("_initialize")),
		handler.GuestConfig(conf),
		handler.Runtime(func(ctx context.Context) (wazero.Runtime, error) {
			return rt, nil
		}))
	if err != nil {
		log.Fatal(err)
	}

	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		for k, strings := range req.Header {
			rw.Header().Set(k, strings[0])
		}
		rw.WriteHeader(http.StatusOK)

	})

	h2 := mw.NewHandler(applyCtx(context.Background()), h)

	h3 := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h2.ServeHTTP(rw, req.WithContext(applyCtx(req.Context())))
	})

	http.DefaultServeMux.Handle("/", h3)

	http.ListenAndServe(":8090", nil)

}

func mockHostModule(runtime wazero.Runtime, guestModule wazero.CompiledModule, modules ...string) {
	for _, module := range modules {
		mb := runtime.NewHostModuleBuilder(module)
		for _, fn := range guestModule.ImportedFunctions() {

			name, mod, _ := fn.Import()

			if name == "rustls_client" {
				mb.NewFunctionBuilder().WithGoFunction(wazeroapi.GoFunc(func(ctx context.Context, stack []uint64) {
					fmt.Println(stack)
				}), fn.ParamTypes(), fn.ResultTypes()).Export(mod)
				fmt.Println(fn.Import())
				fmt.Println(fn.ParamTypes(), fn.ResultTypes())
			}
		}
		mb.Instantiate(context.Background())
	}
}
