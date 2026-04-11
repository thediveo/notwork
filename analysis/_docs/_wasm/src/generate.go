//go:generate bash -c "cp \"$(go env GOROOT)/lib/wasm/wasm_exec.js\" ../../_plugins/go-wasm"
//go:generate ./wasmize.sh

package src
