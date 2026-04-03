/*
Package defers defines an Analyzer that checks for common mistakes in defer
statements using notwork functions such as EnterTransient.

# Analyzer entertransient

entertransient: detect incorrect defers of EnterTransient instead of its result

This checker reports deferred calls to EnterTransient itself, instead of to the
namespace-restoring function returned by EnterTransient.

	defer netns.EnterTransient()    // faulty
	defer netns.EnterTransient()()  // correct
*/
package entertransient
