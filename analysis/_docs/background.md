# Background Details

## SSA Pretty-Printing

## Live Examples

The general idea:

Go program  WASM binary  Docsify WASM Plugin  WASM execution  XTerm.js

Given the following little Go program...

```go
package main

import "fmt"

func main() {
	fmt.Println("The Gophers in your browser say: hellorld!")
}
```

...the resulting (compressed) WASM binary of ~740k is then loaded into the
browser when navigating to this page and executed in a separate [web
worker](https://en.wikipedia.org/wiki/Web_worker) in the background -- to keep
the web page responsive.

<div class='wasm-terminal' data-wasm='hellorld.wasm' data-rows="4"></div>

