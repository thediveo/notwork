// Copyright 2024 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netdevsim

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/devlink"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/netdevsim/ensure"
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"   //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"      //lint:ignore ST1001 rule does not apply
	. "github.com/thediveo/success" //lint:ignore ST1001 rule does not apply
)

// MacvlanPrefix is the name prefix used for transient port network interfaces
// of a transient netdevsim device.
const NetdevsimPrefix = "ndsi-"

var (
	_    = ensure.Netdevsim // ?%$@~*! godoc
	fail = Fail             // allow testing Fails without terminally failing the current test.
)

type Options struct {
	HasID      bool // false means: shut up and get me the next available ID!
	ID         uint
	Ports      uint
	QueueCount uint // per RX and per TX respectively
	NetnsFd    int  // valid when >= 0
	MaxVFs     uint
}

// Opt is a configuration option when creating a new netdevsim network
// interface.
type Opt func(*Options) error

const (
	netdevSimBus          = "netdevsim"
	netdevsimRoot         = "/sys/bus/" + netdevSimBus
	netdevsimDevicesPath  = netdevsimRoot + "/devices"
	netdevsimDevicePrefix = "netdevsim"
)

// HasNetdevsim returns true if netdevsims are available on this host.
//
// Deprecated: use [load.Try] instead that tries to load the netdevsim kernel
// module if not yet loaded.
func HasNetdevsim() bool {
	_, err := os.Stat(netdevsimRoot)
	return err == nil
}

// NewTransient creates a transient netdevsim device as well as at least one
// “port” network interface. The number of port network interfaces and the
// amount of RX+TX queue sets can be specified through options passed in opts.
//
// NewTransient returns the “port” links created, with the first element being
// port 0, the second port 1, and so on. The link objects returned have only
// their [LinkAttrs.Name] set, and optionally their (network)
// [LinkAttrs.Namespace] when configured with the option [InNamespace].
func NewTransient(opts ...Opt) (id uint, links []netlink.Link) {
	GinkgoHelper()

	options := &Options{
		Ports:      1,
		QueueCount: 1,
		NetnsFd:    -1,
	}
	for _, opt := range opts {
		Expect(opt(options)).To(Succeed())
	}

	if options.NetnsFd >= 0 {
		netns.Execute(options.NetnsFd, func() {
			id, links = newTransient(options)
		})
	} else {
		id, links = newTransient(options)
	}
	return
}

// newTransient does the real work of creating a netdevsim device with the given
// configuration options.
//
// Please note that newTransient always creates the netdevsim network interface
// in the current network namespace. So the caller needs to switch to a
// different network namespace where needed.
func newTransient(options *Options) (uint, []netlink.Link) {
	GinkgoHelper()

	// We need a NETLINK devlink API connection in order to query netdevsim
	// device information, such as the mapping of ports to network interface
	// names.
	devlink := Successful(devlink.New())
	defer devlink.Close()

	// Ensure to remove the netdevsim device in case we created one successfully
	// and then failed further down the road, such as when listing and renaming
	// the port network interfaces.
	removeNetdevsim := false
	var id uint
	defer func() {
		// As we're dealing with bus devices, these are not netns-aware, only
		// their port network interfaces are. However, the network interfaces
		// already cease to exist when we remove the netdevsim device. In
		// consequence, we don't need to keep a netns reference to where the
		// interfaces initially appeared, simplifying things.
		if removeNetdevsim {
			_ = os.WriteFile(netdevsimRoot+"/del_device",
				[]byte(strconv.FormatUint(uint64(id), 10)), 0)
		}
	}()

	for attempt := 1; attempt <= 10; attempt++ {
		// locate the "next" available netdevsim ID, unless explicitly specified
		// by caller...
		id = options.ID
		if !options.HasID {
			var err error
			id, err = lowestAvailableID()
			Expect(err).NotTo(HaveOccurred(), "cannot determine available ID")
		}
		By(fmt.Sprintf("creating a transient netdevsim device with ID %d", id))
		// Create the netdevsim device, as well as its ports and thus network
		// interfaces...
		err := os.WriteFile(netdevsimRoot+"/new_device",
			fmt.Appendf(nil, "%d %d %d", id, options.Ports, options.QueueCount), 0)
		if err != nil {
			if options.HasID {
				fail(fmt.Sprintf("cannot create a netdevsim with ID %d, reason: %s",
					id, err.Error()))
			}
			continue // another attempt
		}
		removeNetdevsim = true
		// Wait for the device to appear on the "netdevsim" bus; see also the
		// Linux kernel's netdevsim self tests, such as:
		// https://elixir.bootlin.com/linux/v6.9.6/source/tools/testing/selftests/drivers/net/netdevsim/devlink.sh
		devpath := fmt.Sprintf("%s/%s%d", netdevsimDevicesPath, netdevsimDevicePrefix, id)
		Eventually(func() string { return devpath }).
			Within(2*time.Second).ProbeEvery(1*time.Millisecond).
			Should(BeADirectory(), "netdevsim with ID %d failed to materialize", id)
			// Set the number of VFs
		err = os.WriteFile(devpath+"/sriov_numvfs", []byte(strconv.FormatUint(uint64(options.MaxVFs), 10)), 0)
		if err != nil {
			fail(fmt.Sprintf("cannot set maximum number of %d SR-IOV VFs on netdev with ID %d, reason: %s",
				options.MaxVFs, id, err.Error()))
		}
		// Get the names of the port network interfaces and then rename them using random names.
		nifnames := Successful(portNifnames(devlink, id))
		links := make([]netlink.Link, 0, len(nifnames))
		var netns interface{}
		if options.NetnsFd >= 0 {
			netns = netlink.NsFd(options.NetnsFd)
		}
	nextnif:
		for _, nifname := range nifnames {
			for attempt := 1; attempt <= 10; attempt++ {
				randomname := link.RandomNifname(NetdevsimPrefix)
				// the port network interfaces of netdevsim devices don't have a
				// "kind" as other virtual interfaces like "veth" do, but
				// instead are virtual hardware interfaces; we thus use
				// netlink's Device type instead of GenericDevice.
				if err := netlink.LinkSetName(&netlink.Device{
					LinkAttrs: netlink.LinkAttrs{
						Name: nifname,
					},
				}, randomname); err != nil {
					continue
				}
				links = append(links, &netlink.Device{
					LinkAttrs: netlink.LinkAttrs{
						Name:      randomname,
						Namespace: netns,
					},
				})
				continue nextnif
			}
			fail("too many failed attempts to generate a random port network interface name")
		}
		removeNetdevsim = false
		DeferCleanup(func() {
			By(fmt.Sprintf("removing transient netdevsim with ID %d", id))
			Expect(os.WriteFile(netdevsimRoot+"/del_device",
				[]byte(strconv.FormatUint(uint64(id), 10)), 0)).To(Succeed())
		})
		return id, links
	}
	fail("too many failed attempts to create a transient netdevsim")
	return 0, nil // not reachable
}

// lowestAvailableID returns the lowest available netdevsim ID.
func lowestAvailableID() (uint, error) {
	devsdirf, err := os.Open(netdevsimDevicesPath)
	if err != nil {
		return 0, fmt.Errorf("cannot list existing netdevsim instances, reason: %w", err)
	}
	defer devsdirf.Close()
	devDirEntries, err := devsdirf.ReadDir(-1)
	if err != nil {
		return 0, fmt.Errorf("cannot list existing netdevsim instances, reason: %w", err)
	}
	ids := map[uint]struct{}{}
	for _, devEntry := range devDirEntries {
		name := strings.TrimPrefix(devEntry.Name(), netdevsimDevicePrefix)
		id, err := strconv.ParseUint(name, 10, 32)
		if err != nil {
			continue
		}
		ids[uint(id)] = struct{}{}
	}
	id := uint(0)
	for {
		if _, ok := ids[id]; !ok {
			return id, nil
		}
		id++
	}
}

// portNifnames returns a list of network interface names corresponding with the
// ports of a netdevsim device with the specified ID. The returned name list is
// ordered from port 0 on upwards.
func portNifnames(cl *devlink.Client, id uint) ([]string, error) {
	// the devlink package unfortunately doesn't yet support querying the ports
	// of only a specific device and always dumps all ports.
	ports, err := cl.Ports()
	if err != nil {
		return nil, fmt.Errorf("cannot list netdevsim ports, reason: %w", err)
	}
	devname := netdevsimDevicePrefix + strconv.FormatUint(uint64(id), 10)
	nifnames := []string{}
	for _, port := range ports {
		if port.Bus != netdevSimBus || port.Device != devname {
			continue
		}
		if port.Port >= len(nifnames) {
			newnifnames := make([]string, port.Port+1)
			copy(newnifnames, nifnames)
			nifnames = newnifnames
		}
		nifnames[port.Port] = port.Name
	}
	return nifnames, nil
}
