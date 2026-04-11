# Defer

Let's take a closer look at what happens when using `defer` when using a global
variable `F` that has been initialized with a specific `function`, here `Foo`.
This is the source code we're analyzing:

```go id=defer.go
package main

func Foo() func() {
	return func() { }
}

var F = Foo

func main() {
	defer F()
	defer F()()
}
```

And below is the pretty-printed SSA "graph", dynamically rendered from the
parsed source code above immediately above.

<div class='wasm-terminal' data-wasm='prettyssa.wasm' data-codeid="defer.go" data-rows="85"></div>
