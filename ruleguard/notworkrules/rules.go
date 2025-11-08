package notworkrules

import "github.com/quasilyte/go-ruleguard/dsl"

var Bundle = dsl.Bundle{}

//doc:summary Detects invalid deferred calls to netns.EnterTransient.
//doc:before  defer netns.EnterTransient()
//doc:after   defer netns.EnterTransient()()
func deferredNetnsEnterTransientCall(m dsl.Matcher) { //nolint:unused
	m.Import("github.com/thediveo/notwork/netns")

	m.Match(`defer netns.EnterTransient();`).
		Report("invalid deferred call to netns.EnterTransient itself; instead, defer the result of the call to netns.EnterTransient").
		Suggest(`defer netns.EnterTransient()()`)
}

//doc:summary Detects invalid deferred calls to mntns.EnterTransient.
//doc:before  defer mntns.EnterTransient()
//doc:after   defer mntns.EnterTransient()()
func deferredMntnsEnterTransientCall(m dsl.Matcher) { //nolint:unused
	m.Import("github.com/thediveo/notwork/mntns")

	m.Match(`defer mntns.EnterTransient();`).
		Report("invalid deferred call to mntns.EnterTransient itself; instead, defer the result of the call to mntns.EnterTransient").
		Suggest(`defer mntns.EnterTransient()()`)
}
