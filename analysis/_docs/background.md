# Background Details

## SSA Pretty-Printing

## Live Examples

The general idea:

Go program  WASM binary  Docsify WASM Plugin  WASM execution  XTerm.js

Given the following little Go program...

```go id=hellorld.go
package main

import (
	"fmt"
	"os"
)

func main() {
	song := "wot?"
	if len(os.Args) > 1 {
		song = os.Args[1]
	}
	fmt.Printf("The 66 dancing Gophers in your browser sing: %s\n", song)
}
```

...the resulting (compressed) WASM binary of ~740k is then loaded into the
browser when this page gets rendered, and then executed in a separate [web
worker](https://en.wikipedia.org/wiki/Web_worker) in the background -- to keep
the web page responsive.

<div class='wasm-terminal' data-wasm='hellorld.wasm' data-args="Hellorld!" data-rows="4"></div>

