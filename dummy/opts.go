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

package dummy

import (
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"
)

// InNamespace configures a dummy network interface to be created in the
// network namespace referenced by fdref, instead of creating it in the current
// network namespace.
func InNamespace(fdref int) Opt {
	return func(l *link.Link) error {
		l.Attrs().Namespace = netlink.NsFd(fdref)
		return nil
	}
}
