/***
Copyright 2014 Cisco Systems Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mastercfg

import (
	"encoding/json"
	"fmt"
	"github.com/contiv/netplugin/core"
	"sync"
)

const (
	vnfConfigPathPrefix       = StateConfigPath + "vnf/"
	vnfConfigPath             = vnfConfigPathPrefix + "%s"
	vnfPolicyConfigPathPrefix = StateConfigPath + "vnfpolicy/"
	vnfPolicyConfigPath       = vnfConfigPathPrefix + "%s"
)

// VnfInfo holds service information
type VnfInfo struct {
	VnfName   string            // VNF name
	Tenant    string            // Tenant name
	Group     string            // VNF network
	VnfLabels map[string]string // VNF labels associated with a VNF
	//	VnfInstances map[string]*VnfInstance // map of providers for a service keyed by provider ip
}

// VnfLBDb is map of all VNFs
var VnfLBDb = make(map[string]*VnfInfo) // DB for all VNFs keyed by vnf.tenant

// VnfMutex is mutex for vnf transaction
var VnfMutex sync.RWMutex

// CfgVnfState is the service object configuration
type CfgVnfState struct {
	core.CommonState
	VnfName   string            `json:"vnfName"`
	Tenant    string            `json:"tenantName"`
	Group     string            `json:"group"`
	VnfLabels map[string]string `json:"vnfLabels"`
	VtepIP    string            `json:"vtepIP"`
	//	VnfInstances map[string]*VnfInstance `json:"vnfInstances"`
}

// Write the state
func (s *CfgVnfState) Write() error {
	key := fmt.Sprintf(vnfConfigPath, s.ID)
	err := s.StateDriver.WriteState(key, s, json.Marshal)
	return err
}

// Read the state in for a given ID.
func (s *CfgVnfState) Read(id string) error {
	key := fmt.Sprintf(vnfConfigPath, id)
	err := s.StateDriver.ReadState(key, s, json.Unmarshal)
	return err
}

// ReadAll reads all the state for master bgp configurations and returns it.
func (s *CfgVnfState) ReadAll() ([]core.State, error) {
	return s.StateDriver.ReadAllState(vnfConfigPathPrefix, s, json.Unmarshal)
}

// Clear removes the configuration from the state store.
func (s *CfgVnfState) Clear() error {
	key := fmt.Sprintf(vnfConfigPath, s.ID)
	err := s.StateDriver.ClearState(key)
	return err
}

// WatchAll state transitions and send them through the channel.
func (s *CfgVnfState) WatchAll(rsps chan core.WatchState) error {
	return s.StateDriver.WatchAllState(vnfConfigPathPrefix, s, json.Unmarshal,
		rsps)
}
