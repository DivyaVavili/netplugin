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
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/contiv/contivmodel"
	"github.com/contiv/netplugin/core"
	"github.com/contiv/ofnet"
)

const (
	vnfPolicyConfigPathPrefix = StateConfigPath + "vnfpolicy/"
	vnfPolicyConfigPath       = vnfPolicyConfigPathPrefix + "%s"
)

// VnfPolicyState has an instance of policy attached to an endpoint group
type VnfPolicyState struct {
	core.CommonState
	VnfPolicyKey string                 // Key for this VNF policy
	SrcUnitID    int                    // Src Unit
	DestUnitID   int                    // Dest Unit
	AttachedVNF  *contivModel.Vnf       // Attached VNF
	VnfOfnetRule *ofnet.OfnetPolicyRule // Ofnet Rule associated with this policy
}

// VnfPolicyDb database
var VnfPolicyDb = make(map[string]*VnfPolicyState)

// Create the netmaster
var vnfOfnetMaster *ofnet.OfnetMaster

// state store
var vnfStateStore core.StateDriver

// InitVnfPolicyMgr initializes the policy manager
func InitVnfPolicyMgr(stateDriver core.StateDriver, ofm *ofnet.OfnetMaster) error {
	// save statestore and ofnet masters
	vnfStateStore = stateDriver
	vnfOfnetMaster = ofm

	// restore all existing VNF policies
	err := restoreVnfPolicies(stateDriver)
	if err != nil {
		log.Errorf("Error restoring VNF policies. ")
	}
	return nil
}

// InstallVnfPolicy creates a new policy instance attached to an endpoint group
func InstallVnfPolicy(policyID string, srcUnitID, destUnitID int, vnf *contivModel.Vnf) error {
	vnfPolicyKey := GetVnfPolicyID(srcUnitID, destUnitID, vnf)

	// See if it already exists
	vnfp := FindVnfPolicy(vnfPolicyKey)
	if vnfp != nil {
		log.Errorf("VNF policy %s already exists", vnfPolicyKey)
		return core.Errorf("VNF policy exists")
	}

	vPolicy := new(VnfPolicyState)
	vPolicy.VnfPolicyKey = vnfPolicyKey
	vPolicy.ID = policyID
	vPolicy.SrcUnitID = srcUnitID
	vPolicy.DestUnitID = destUnitID
	vPolicy.AttachedVNF = vnf
	vPolicy.StateDriver = vnfStateStore

	log.Infof("Creating new VNF policy: %s", vnfPolicyKey)

	// TODO: Install Ofnet Rule

	// Save the policy state
	err := vPolicy.Write()
	if err != nil {
		return err
	}

	// Save it in local cache
	VnfPolicyDb[vnfPolicyKey] = vPolicy

	log.Info("Created VNF policy %+v", vPolicy)

	return nil
}

// UninstallVnfPolicy creates a new policy instance attached to an endpoint group
func UninstallVnfPolicy(policyID string, srcUnitID, destUnitID int, vnf *contivModel.Vnf) error {
	vnfPolicyKey := GetVnfPolicyID(srcUnitID, destUnitID, vnf)

	// See if it already exists
	vnfp := FindVnfPolicy(vnfPolicyKey)
	if vnfp == nil {
		log.Errorf("VNF policy %s does not exist", vnfPolicyKey)
		return core.Errorf("VNF policy does not exist")
	}

	// TODO: Uninstall Ofnet Rule

	// Delete the policy state
	err := vnfp.Delete()
	if err != nil {
		return err
	}

	log.Info("Deleted VNF policy %+v", vnfp)

	return nil
}

// restoreVnfPolicies restores all VNF policies from state store
func restoreVnfPolicies(stateDriver core.StateDriver) error {
	// read all VNF policies
	vPolicy := new(VnfPolicyState)
	vPolicy.StateDriver = stateDriver
	vnfPolicyCfgs, err := vPolicy.ReadAll()
	if err == nil {
		for _, vnfPolicyCfg := range vnfPolicyCfgs {
			vnfp := vnfPolicyCfg.(*VnfPolicyState)
			log.Infof("Restoring VnfPolicy: %+v", vnfp)

			// save it in cache
			VnfPolicyDb[vPolicy.VnfPolicyKey] = vnfp

			// TODO: Install necessary rules
		}
	}

	return nil
}

// FindVnfPolicy finds an VNF policy
func FindVnfPolicy(policyKey string) *VnfPolicyState {
	return VnfPolicyDb[policyKey]
}

// Delete deletes the VNF policy
func (vPolicy *VnfPolicyState) Delete() error {

	log.Infof("Before deleting VNFPolicyDb: %+v", VnfPolicyDb)
	// delete from the DB
	delete(VnfPolicyDb, vPolicy.VnfPolicyKey)

	log.Infof("After deleting VNFPolicyDb: %+v", VnfPolicyDb)
	return vPolicy.Clear()
}

// GetVnfPolicyID returns the VNF policy ID for state lookup
func GetVnfPolicyID(srcUnitID, destUnitID int, vnf *contivModel.Vnf) string {
	return vnf.TenantName + ":" + strconv.Itoa(srcUnitID) + ":" + strconv.Itoa(destUnitID) + ":" + strings.Split(vnf.Key, ":")[1]
}

/* createOfnetRule creates a directional ofnet rule
func (gp *VnfPolicyState) createOfnetRule(rule *contivModel.Rule, dir string) (*ofnet.OfnetPolicyRule, error) {
	var remoteEpgID int
	var err error

	ruleID := vPolicy.VnfPolicyKey + ":" + rule.Key + ":" + dir

	// Create an ofnet rule
	ofnetRule := new(ofnet.OfnetPolicyRule)
	ofnetRule.RuleId = ruleID
	ofnetRule.Priority = rule.Priority
	ofnetRule.Action = rule.Action

	// See if user specified an endpoint Group in the rule
	if rule.FromEndpointGroup != "" {
		remoteEpgID, err = GetEndpointGroupID(vnfStateStore, rule.FromEndpointGroup, rule.TenantName)
		if err != nil {
			log.Errorf("Error finding endpoint group %s/%s/%s. Err: %v",
				rule.FromEndpointGroup, rule.FromNetwork, rule.TenantName, err)
		}
	} else if rule.ToEndpointGroup != "" {
		remoteEpgID, err = GetEndpointGroupID(vnfStateStore, rule.ToEndpointGroup, rule.TenantName)
		if err != nil {
			log.Errorf("Error finding endpoint group %s/%s/%s. Err: %v",
				rule.ToEndpointGroup, rule.ToNetwork, rule.TenantName, err)
		}
	} else if rule.FromNetwork != "" {
		netKey := rule.TenantName + ":" + rule.FromNetwork

		net := contivModel.FindNetwork(netKey)
		if net == nil {
			log.Errorf("Network %s not found", netKey)
			return nil, errors.New("FromNetwork not found")
		}

		rule.FromIpAddress = net.Subnet
	} else if rule.ToNetwork != "" {
		netKey := rule.TenantName + ":" + rule.ToNetwork

		net := contivModel.FindNetwork(netKey)
		if net == nil {
			log.Errorf("Network %s not found", netKey)
			return nil, errors.New("ToNetwork not found")
		}

		rule.ToIpAddress = net.Subnet
	}

	// Set protocol
	switch rule.Protocol {
	case "tcp":
		ofnetRule.IpProtocol = 6
	case "udp":
		ofnetRule.IpProtocol = 17
	case "icmp":
		ofnetRule.IpProtocol = 1
	case "igmp":
		ofnetRule.IpProtocol = 2
	case "":
		ofnetRule.IpProtocol = 0
	default:
		proto, err := strconv.Atoi(rule.Protocol)
		if err == nil && proto < 256 {
			ofnetRule.IpProtocol = uint8(proto)
		}
	}

	// Set directional parameters
	switch dir {
	case "inRx":
		// Set src/dest endpoint group
		ofnetRule.DstEndpointGroup = vPolicy.EndpointGroupID
		ofnetRule.SrcEndpointGroup = remoteEpgID

		// Set src/dest IP Address
		ofnetRule.SrcIpAddr = rule.FromIpAddress

		// set port numbers
		ofnetRule.DstPort = uint16(rule.Port)

		// set tcp flags
		if rule.Protocol == "tcp" && rule.Port == 0 {
			ofnetRule.TcpFlags = "syn,!ack"
		}
	case "inTx":
		// Set src/dest endpoint group
		ofnetRule.SrcEndpointGroup = vPolicy.EndpointGroupID
		ofnetRule.DstEndpointGroup = remoteEpgID

		// Set src/dest IP Address
		ofnetRule.DstIpAddr = rule.FromIpAddress

		// set port numbers
		ofnetRule.SrcPort = uint16(rule.Port)
	case "outRx":
		// Set src/dest endpoint group
		ofnetRule.DstEndpointGroup = vPolicy.EndpointGroupID
		ofnetRule.SrcEndpointGroup = remoteEpgID

		// Set src/dest IP Address
		ofnetRule.SrcIpAddr = rule.ToIpAddress

		// set port numbers
		ofnetRule.SrcPort = uint16(rule.Port)
	case "outTx":
		// Set src/dest endpoint group
		ofnetRule.SrcEndpointGroup = vPolicy.EndpointGroupID
		ofnetRule.DstEndpointGroup = remoteEpgID

		// Set src/dest IP Address
		ofnetRule.DstIpAddr = rule.ToIpAddress

		// set port numbers
		ofnetRule.DstPort = uint16(rule.Port)

		// set tcp flags
		if rule.Protocol == "tcp" && rule.Port == 0 {
			ofnetRule.TcpFlags = "syn,!ack"
		}
	default:
		log.Fatalf("Unknown rule direction %s", dir)
	}

	// Add the Rule to policyDB
	err = vnfOfnetMaster.AddRule(ofnetRule)
	if err != nil {
		log.Errorf("Error creating rule {%+v}. Err: %v", ofnetRule, err)
		return nil, err
	}

	log.Infof("Added rule {%+v} to policyDB", ofnetRule)

	return ofnetRule, nil
}

// AddRule adds a rule to epg policy
func (gp *VnfPolicyState) AddRule(rule *contivModel.Rule) error {
	var dirs []string

	// check if the rule exists already
	if vPolicy.RuleMaps[rule.Key] != nil {
		// FIXME: see if we can update the rule
		return core.Errorf("Rule already exists")
	}

	// Figure out all the directional rules we need to install
	switch rule.Direction {
	case "in":
		if (rule.Protocol == "udp" || rule.Protocol == "tcp") && rule.Port != 0 {
			dirs = []string{"inRx", "inTx"}
		} else {
			dirs = []string{"inRx"}
		}
	case "out":
		if (rule.Protocol == "udp" || rule.Protocol == "tcp") && rule.Port != 0 {
			dirs = []string{"outRx", "outTx"}
		} else {
			dirs = []string{"outTx"}
		}
	case "both":
		if (rule.Protocol == "udp" || rule.Protocol == "tcp") && rule.Port != 0 {
			dirs = []string{"inRx", "inTx", "outRx", "outTx"}
		} else {
			dirs = []string{"inRx", "outTx"}
		}

	}

	// create a ruleMap
	ruleMap := new(RuleMap)
	ruleMap.OfnetRules = make(map[string]*ofnet.OfnetPolicyRule)
	ruleMap.Rule = rule

	// Create ofnet rules
	for _, dir := range dirs {
		ofnetRule, err := vPolicy.createOfnetRule(rule, dir)
		if err != nil {
			log.Errorf("Error creating %s ofnet rule for {%+v}. Err: %v", dir, rule, err)
			return err
		}

		// add it to the rule map
		ruleMap.OfnetRules[ofnetRule.RuleId] = ofnetRule
	}

	// save the rulemap
	vPolicy.RuleMaps[rule.Key] = ruleMap

	return nil
}

// DelRule removes a rule from VNF policy
func (gp *VnfPolicyState) DelRule(rule *contivModel.Rule) error {
	// check if the rule exists
	ruleMap := vPolicy.RuleMaps[rule.Key]
	if ruleMap == nil {
		return core.Errorf("Rule does not exists")
	}

	// Delete each ofnet rule under this policy rule
	for _, ofnetRule := range ruleMap.OfnetRules {
		log.Infof("Deleting rule {%+v} from policyDB", ofnetRule)

		// Delete the rule from policyDB
		err := vnfOfnetMaster.DelRule(ofnetRule)
		if err != nil {
			log.Errorf("Error deleting the ofnet rule {%+v}. Err: %v", ofnetRule, err)
		}
	}

	// delete the cache
	delete(vPolicy.RuleMaps, rule.Key)

	return nil
}
*/
// Write the state.
func (vPolicy *VnfPolicyState) Write() error {
	key := fmt.Sprintf(vnfPolicyConfigPath, vPolicy.ID)
	return vPolicy.StateDriver.WriteState(key, vPolicy, json.Marshal)
}

// Read the state for a given identifier
func (vPolicy *VnfPolicyState) Read(id string) error {
	key := fmt.Sprintf(vnfPolicyConfigPath, id)
	return vPolicy.StateDriver.ReadState(key, vPolicy, json.Unmarshal)
}

// ReadAll state and return the collection.
func (vPolicy *VnfPolicyState) ReadAll() ([]core.State, error) {
	return vPolicy.StateDriver.ReadAllState(vnfPolicyConfigPathPrefix, vPolicy, json.Unmarshal)
}

// WatchAll state transitions and send them through the channel.
func (vPolicy *VnfPolicyState) WatchAll(rsps chan core.WatchState) error {
	return vPolicy.StateDriver.WatchAllState(vnfPolicyConfigPathPrefix, vPolicy, json.Unmarshal,
		rsps)
}

// Clear removes the state.
func (vPolicy *VnfPolicyState) Clear() error {
	key := fmt.Sprintf(vnfPolicyConfigPath, vPolicy.ID)
	return vPolicy.StateDriver.ClearState(key)
}
