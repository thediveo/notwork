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
	"errors"
	"fmt"
)

// WithID configures a new netdevsim to use the specified ID, as opposed to the
// lowest available ID.
func WithID(id uint) Opt {
	return func(o *Options) error {
		o.HasID = true
		o.ID = id
		return nil
	}
}

// WithPorts configures a new netdevsim to have the specified number of ports
// (=individual network interfaces).
func WithPorts(n uint) Opt {
	return func(o *Options) error {
		o.Ports = n
		return nil
	}
}

// WithRxTxQueueCountEach configures a new netdevsim to have the specified
// number of RX as well as TX queues. Specifying a zero queue count results in
// an error when trying to create a netdevsim.
func WithRxTxQueueCountEach(n uint) Opt {
	return func(o *Options) error {
		if n == 0 {
			return errors.New("RX/TX queue count cannot be zero")
		}
		o.QueueCount = n
		return nil
	}
}

// WithMaxVFs configures a new netdevsim to have the specified number of VFs.
// Specifying zero means “no VFs”.
func WithMaxVFs(n uint) Opt {
	return func(o *Options) error {
		o.MaxVFs = n
		return nil
	}
}

// InNamespace configures a new netdevsim to have its port network interface(s)
// to be created in the network namespace referenced by fdref, instead of
// creating it in the current network namespace.
func InNamespace(fdref int) Opt {
	return func(o *Options) error {
		if fdref < 0 {
			return fmt.Errorf("invalid netns fd %d", fdref)
		}
		o.NetnsFd = fdref
		return nil
	}
}
