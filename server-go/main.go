package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/http-wasm/http-wasm-host-go/api"
	"github.com/http-wasm/http-wasm-host-go/handler"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
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
	"CONF":"VAL",
	"X-Field":"Value"cu
}}`)

	mw, err := wasm.NewMiddleware(context.Background(), code, handler.Logger(logger{}), handler.GuestConfig(conf))
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

	http.ListenAndServe(":8090", h2)

}
