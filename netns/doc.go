/*
Package netns supports working with network namespace IDs (“nsid”) and netlink
handles in unit tests.

For handling network namespaces and their identifiers in general, please refer
to the github.com/thediveo/spacetest/netns package instead. The (deprecated)
test helper functions in this package now refer to their twins from the new
package. Development and maintenance of general network namespace-related
functionality from now on will be only on the “spacetest” module, which has the
benefit of not coming with any netlink-related dependencies. Instead, any
netlink-related dependencies are kept with the “notwork” module.
*/
package netns
