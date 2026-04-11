# Simple Main

Let's analyse the [SSA IR (intermediate
representation)](https://pkg.go.dev/golang.org/x/tools/go/ssa) for this little
program below:

```go id=bar
package main

func main() {
	println("Hellorld!")
}
```

Following is our _live rendering_ of the SSA data for this example using my own
stupid pretty-printer for SSA package graphs -- for this, the above code has
been compiled into [WASM](https://go.dev/wiki/WebAssembly):

<div class='wasm-terminal' data-wasm='simplemain.wasm' data-rows="30"></div>
