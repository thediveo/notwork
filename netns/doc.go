/*
Package netns supports running unit tests in separated transient network
namespaces.

# Usage

Just call [EnterTransientNetns] and defer its return value.

	  import "github.com/notwork/netns"

	  It("tests something inside a temporary network namespace", func() {
		defer netns.EnterTransientNetns()
		// ...
	  })
*/
package netns
