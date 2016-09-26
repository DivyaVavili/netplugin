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

package master

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel"
	"github.com/contiv/netplugin/netmaster/mastercfg"
	"github.com/contiv/netplugin/utils"
)

// CreateVnfPolicy creates VNF policy
func CreateVnfPolicy(vnfPolicy *contivModel.VnfPolicy) error {
	log.Infof("Received VNF policy create")

	// Skip policy insertions in ACI mode
	if !isPolicyEnabled() {
		return nil
	}

	// Check for valid EPGs and VNF
	srcEpgKey := vnfPolicy.TenantName + ":" + vnfPolicy.SourceUnit
	srcEpg := contivModel.FindEndpointGroup(srcEpgKey)
	if srcEpg == nil {
		err := fmt.Errorf("Src EPG %s not found during VNF policy create", srcEpgKey)
		return err
	}

	destEpgKey := vnfPolicy.TenantName + ":" + vnfPolicy.DestUnit
	destEpg := contivModel.FindEndpointGroup(destEpgKey)
	if destEpg == nil {
		err := fmt.Errorf("Dest EPG %s not found during VNF policy create", destEpgKey)
		return err
	}

	vnfKey := vnfPolicy.TenantName + ":" + vnfPolicy.Vnf
	vnf := contivModel.FindVnf(vnfKey)
	if vnf == nil {
		err := fmt.Errorf("Vnf %s not found during VNF policy create", vnfKey)
		return err
	}

	stateDriver, err := utils.GetStateDriver()
	if err != nil {
		log.Errorf("Could not get StateDriver during VNF policy create {%s}", vnfPolicy.Key)
		return err
	}

	srcEpgID, err := mastercfg.GetEndpointGroupID(stateDriver, srcEpg.GroupName, srcEpg.TenantName)
	if err != nil {
		log.Errorf("Error getting epgID for %+v during VNG policy create. Err: %v", srcEpg, err)
		return err
	}

	destEpgID, err := mastercfg.GetEndpointGroupID(stateDriver, destEpg.GroupName, destEpg.TenantName)
	if err != nil {
		log.Errorf("Error getting epgID for %+v during VNG policy create. Err: %v", destEpg, err)
		return err
	}

	// Install the VNF policy
	err = mastercfg.InstallVnfPolicy(vnfPolicy.Key, srcEpgID, destEpgID, vnf)
	if err != nil {
		log.Errorf("Error creating VNF policy. Err: %v", err)
		return err
	}

	return nil
}

// DeleteVnfPolicy deletes VNF policy
func DeleteVnfPolicy(vnfPolicy *contivModel.VnfPolicy) error {
	log.Infof("Received Vnf Policy delete in master. TBD")

	// Check for valid EPGs and VNF
	srcEpgKey := vnfPolicy.TenantName + ":" + vnfPolicy.SourceUnit
	srcEpg := contivModel.FindEndpointGroup(srcEpgKey)
	if srcEpg == nil {
		err := fmt.Errorf("Src EPG %s not found during VNF policy delete", srcEpgKey)
		return err
	}

	destEpgKey := vnfPolicy.TenantName + ":" + vnfPolicy.DestUnit
	destEpg := contivModel.FindEndpointGroup(destEpgKey)
	if destEpg == nil {
		err := fmt.Errorf("Dest EPG %s not found during VNF policy delete", destEpgKey)
		return err
	}

	vnfKey := vnfPolicy.TenantName + ":" + vnfPolicy.Vnf
	vnf := contivModel.FindVnf(vnfKey)
	if vnf == nil {
		err := fmt.Errorf("Vnf %s not found during VNF policy delete", vnfKey)
		return err
	}

	stateDriver, err := utils.GetStateDriver()
	if err != nil {
		log.Errorf("Could not get StateDriver during VNF policy delete {%s}", vnfPolicy.Key)
		return err
	}

	srcEpgID, err := mastercfg.GetEndpointGroupID(stateDriver, srcEpg.GroupName, srcEpg.TenantName)
	if err != nil {
		log.Errorf("Error getting epgID for %+v during VNG policy delete. Err: %v", srcEpg, err)
		return err
	}

	destEpgID, err := mastercfg.GetEndpointGroupID(stateDriver, destEpg.GroupName, destEpg.TenantName)
	if err != nil {
		log.Errorf("Error getting epgID for %+v during VNG policy delete. Err: %v", destEpg, err)
		return err
	}

	// Uninstall the VNF policy
	err = mastercfg.UninstallVnfPolicy(vnfPolicy.Key, srcEpgID, destEpgID, vnf)
	if err != nil {
		log.Errorf("Error creating VNF policy. Err: %v", err)
		return err
	}

	return nil
}
