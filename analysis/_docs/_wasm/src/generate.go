//go:generate bash -c "cp \"$(go env GOROOT)/lib/wasm/wasm_exec.js\" ../../_plugins/gowasm/"
//go:generate ./wasmize.sh

package src
