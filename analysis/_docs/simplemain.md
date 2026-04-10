# Simple Main

Let's analyse the SSA graph for this little program...

```go
package main

func main() {
	println("Hellorld!")
}
```

Following is a live rendering of this example -- that has been compiled into
[WASM](https://go.dev/wiki/WebAssembly) -- using my own stupid pretty-printer
for SSA package graphs:

<div class='wasm-terminal' data-wasm='simplemain.wasm' data-rows="35"></div>
