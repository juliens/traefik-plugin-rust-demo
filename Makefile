target/wasm32-wasi/release/http-wasm-header-plugin.wasm: ./src/*.rs
	cargo build --target wasm32-wasip1

target/wasm32-wasip1/debug/http-wasm-header-plugin.wasm: ./src/*.rs
	cargo build --target wasm32-wasip1 # --release

plugin.wasm: target/wasm32-wasip1/debug/http-wasm-header-plugin.wasm
	cp ./target/wasm32-wasip1/debug/http-wasm-header-plugin.wasm ./plugin.wasm

.PHONY=build
build: plugin.wasm

build-release:
	cargo build --target wasm32-wasip1 --release
	cp target/wasm32-wasi/release/http-wasm-header-plugin.wasm ./plugin.wasm
	
clean:
	rm -rf target plugin.wasm
