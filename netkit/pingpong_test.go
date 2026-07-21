// Copyright 2026 Harald Albrecht.
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

package netkit

import (
	"net"
	"os"
	"syscall"
	"time"

	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient netkit network interface pairs", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	DescribeTable("pings from peer to peer",
		func(primpol, peerpol netlink.NetkitPolicy, pings bool) {
			primaryIP := Successful(netlink.ParseAddr("192.0.2.1/24"))
			peerIP := Successful(netlink.ParseAddr("192.0.2.2/24"))

			defer netns.EnterTransient()()      // Dupond...
			peernetnsfd := netns.NewTransient() // ...et Dupont

			By("creating a pair of netkits, bringing them up, and setting IP addresses")
			primary, peer := NewTransient(WithL2Mode(),
				WithPolicy(primpol),
				WithPeerNamespace(peernetnsfd),
				WithPeerPolicy(peerpol))
			Expect(netlink.LinkByName(peer.Attrs().Name)).Error().To(HaveOccurred(),
				"WithPeerNamespace failed")
			netns.Execute(peernetnsfd, func() {
				Expect(netlink.LinkByName(peer.Attrs().Name)).Error().NotTo(HaveOccurred())
				Expect(netlink.AddrAdd(peer, peerIP)).To(Succeed())
				link.Up(peer)
			})

			Expect(netlink.AddrAdd(primary, primaryIP)).To(Succeed())
			link.Up(primary)
			link.EnsureUp(primary)

			By("pinging the peer")
			conn := Successful(icmp.ListenPacket(
				"ip4:icmp", primaryIP.IP.String()))
			DeferCleanup(conn.Close)

			echo := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &icmp.Echo{
					ID:   42,
					Seq:  1,
					Data: []byte("HELO"),
				},
			}
			Expect(conn.WriteTo(Successful(
				echo.Marshal(nil)),
				&net.IPAddr{IP: net.ParseIP(peerIP.IP.String())})).
				To(Equal(1 + 1 + (2) + 2 + 2 + 4))

			By("receiving the echo")
			reply := make([]byte, 60+8+576)
			Expect(conn.SetDeadline(time.Now().Add(1 * time.Second))).To(Succeed())
			if pings {
				n, _ := Successful2R(conn.ReadFrom(reply))
				echoreply := Successful(icmp.ParseMessage(syscall.IPPROTO_ICMP, reply[:n]))
				body := AssignableTo[*icmp.Echo](echoreply.Body)
				Expect(body.Data).To(Equal([]byte("HELO")))
				return
			}
			Expect(conn.ReadFrom(reply)).Error().To(MatchError(ContainSubstring("i/o timeout")))
		},
		Entry(nil, netlink.NETKIT_POLICY_FORWARD, netlink.NETKIT_POLICY_FORWARD, true),
		Entry(nil, netlink.NETKIT_POLICY_BLACKHOLE, netlink.NETKIT_POLICY_FORWARD, false),
		Entry(nil, netlink.NETKIT_POLICY_FORWARD, netlink.NETKIT_POLICY_BLACKHOLE, false),
	)

})
