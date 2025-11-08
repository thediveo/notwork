package notwork

import (
	"github.com/thediveo/notwork/mntns"
	"github.com/thediveo/notwork/netns"
)

func foo() {
	defer netns.EnterTransient()
	defer mntns.EnterTransient()
}
