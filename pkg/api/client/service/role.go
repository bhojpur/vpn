package service

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"strings"

	"github.com/ipfs/go-log"
)

// Role is a service role.
// It is identified by a unique string which is sent over the wire
// and streamed to/from the clients.
// Roles can be applied either directly, or assigned within roles in the API
type Role string

// RoleConfig is the role config structure, which holds all the objects that can be used by a Role
type RoleConfig struct {
	Client                                              *Client
	UUID, ServiceID, StateDir, APIAddress, NetworkToken string
	Logger                                              log.StandardLogger

	roles map[Role]func(c *RoleConfig) error
}

// RoleOption is a role option
type RoleOption func(c *RoleConfig)

// RoleKey is an association between a Role(string) and a Handler which actually
// fullfills the role
type RoleKey struct {
	RoleHandler func(c *RoleConfig) error
	Role        Role
}

// WithRole sets the available roles
func WithRole(f map[Role]func(c *RoleConfig) error) RoleOption {
	return func(c *RoleConfig) {
		c.roles = f
	}
}

// WithRoleLogger sets a logger for the role action
func WithRoleLogger(l log.StandardLogger) RoleOption {
	return func(c *RoleConfig) {
		c.Logger = l
	}
}

// WithRoleUUID sets the UUID which performs the role
func WithRoleUUID(u string) RoleOption {
	return func(c *RoleConfig) {
		c.UUID = u
	}
}

// WithRoleStateDir sets the statedir for the role
func WithRoleStateDir(s string) RoleOption {
	return func(c *RoleConfig) {
		c.StateDir = s
	}
}

// WithRoleToken sets the network token which can be used by the role
func WithRoleToken(s string) RoleOption {
	return func(c *RoleConfig) {
		c.NetworkToken = s
	}
}

// WithRoleAPIAddress sets the API Address used during the execution
func WithRoleAPIAddress(s string) RoleOption {
	return func(c *RoleConfig) {
		c.APIAddress = s
	}
}

// WithRoleServiceID sets a role service ID
func WithRoleServiceID(s string) RoleOption {
	return func(c *RoleConfig) {
		c.ServiceID = s
	}
}

// WithRoleClient sets a client for a role
func WithRoleClient(e *Client) RoleOption {
	return func(c *RoleConfig) {
		c.Client = e
	}
}

// Apply applies a role and takes a list of options
func (rr Role) Apply(opts ...RoleOption) {
	c := &RoleConfig{}
	for _, o := range opts {
		o(c)
	}

	for _, role := range strings.Split(string(rr), ",") {
		r := Role(role)
		if f, exists := c.roles[r]; exists {
			c.Logger.Info("Role loaded. Applying ", r)
			if err := f(c); err != nil {
				c.Logger.Warning("Failed applying role", role, err)
			}
		} else {
			c.Logger.Warn("Unknown role: ", r)
		}
	}
}
