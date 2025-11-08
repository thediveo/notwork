package notwork

import (
	bar "github.com/thediveo/notwork/mntns"
	foo "github.com/thediveo/notwork/netns"
)

func test() {
	defer foo.EnterTransient()()
	defer foo.EnterTransient()
	defer bar.EnterTransient()
}
