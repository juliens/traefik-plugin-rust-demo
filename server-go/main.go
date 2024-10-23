package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	code, err := os.ReadFile("../plugin.wasm")
	if err != nil {
		log.Fatal(err)
	}

	conf := []byte(`{"headers":{
	"X-Field":"Value"
}}`)

	var applyCtx func(context.Context) context.Context
	applyCtx = func(ctx context.Context) context.Context {
		return ctx
	}
	mw, err := wasm.NewMiddleware(context.Background(), code, handler.Logger(logger{}), handler.GuestConfig(conf), handler.Runtime(func(ctx context.Context) (wazero.Runtime, error) {
		rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
		guestModule, err := rt.CompileModule(ctx, code)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Runtime")
		mb := rt.NewHostModuleBuilder("rustls_client")

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
		if extension := imports.DetectSocketsExtension(guestModule); extension != nil {
			builder := imports.NewBuilder().WithSocketsExtension("auto", guestModule)
			ctx, sys, err := builder.Instantiate(ctx, rt)
			if err != nil {
				return nil, err
			}

			builder.WithDirs("/:/")
			inst, err := wazergo.Instantiate(ctx, rt, wazergo_wasip1.NewHostModule(*extension), wazergo_wasip1.WithWASI(sys))
			if err != nil {
				return nil, fmt.Errorf("wazergo instantiation: %w", err)
			}
			applyCtx = func(ctx context.Context) context.Context {
				return wazergo.WithModuleInstance(ctx, inst)
			}
		}

		rt.NewHostModuleBuilder("wasmedge_httpsreq").
			NewFunctionBuilder().WithGoFunction(wazeroapi.GoFunc(func(ctx context.Context, stack []uint64) {
			fmt.Println(stack)
		}), []wazeroapi.ValueType{127, 127, 127, 127, 127}, []wazeroapi.ValueType{}).Export("wasmedge_httpsreq_send_data").
			NewFunctionBuilder().WithGoFunction(wazeroapi.GoFunc(func(ctx context.Context, stack []uint64) {
			fmt.Println(stack)
		}), []wazeroapi.ValueType{127}, []wazeroapi.ValueType{}).Export("wasmedge_httpsreq_get_rcv").
			NewFunctionBuilder().WithGoFunction(wazeroapi.GoFunc(func(ctx context.Context, stack []uint64) {
			fmt.Println(stack)
		}), []wazeroapi.ValueType{}, []wazeroapi.ValueType{127}).Export("wasmedge_httpsreq_get_rcv_len").Instantiate(context.Background())

		return rt, nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		for k, strings := range req.Header {
			fmt.Println(k, ":", strings)
		}

	})

	h2 := mw.NewHandler(context.Background(), h)

	h3 := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h2.ServeHTTP(rw, req.WithContext(applyCtx(req.Context())))
	})

	http.ListenAndServe(":8090", h3)

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
