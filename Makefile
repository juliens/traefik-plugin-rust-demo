target/wasm32-wasi/release/http-wasm-header-plugin.wasm: ./src/*.rs
	cargo build --target wasm32-wasi --release

plugin.wasm: target/wasm32-wasi/release/http-wasm-header-plugin.wasm
	cp ./target/wasm32-wasi/release/http-wasm-header-plugin.wasm ./plugin.wasm

.PHONY=build
build: plugin.wasm
	

