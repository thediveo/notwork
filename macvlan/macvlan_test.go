package macvlan

import (
	"os"

	"github.com/thediveo/notwork/dummy"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("provides transient MACVLAN network interfaces", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	It("creates a MACVLAN with a dummy parent", func() {
		_ = CreateTransient(dummy.NewTransientUp())
	})

	It("finds a hardware NIC in up state", func() {
		parent := LocateHWParent()
		Expect(parent).NotTo(BeNil())
	})

})
