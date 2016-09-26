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
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/netplugin/netmaster/intent"
	"github.com/contiv/netplugin/netmaster/mastercfg"
	"github.com/contiv/netplugin/utils"
)

// CreateVNF adds to the etcd state
func CreateVNF(vnfCfg *intent.ConfigVNF) error {

	log.Infof("Received create VNF config {%v}", vnfCfg)

	// Get the state driver
	stateDriver, err := utils.GetStateDriver()
	if err != nil {
		return err
	}

	vnfState := &mastercfg.CfgVnfState{}
	vnfState.ID = GetVnfID(vnfCfg.TenantName, vnfCfg.VnfName)
	vnfState.StateDriver = stateDriver
	vnfState.VnfName = vnfCfg.VnfName
	vnfState.Tenant = vnfCfg.TenantName
	vnfState.TrafficAction = vnfCfg.TrafficAction
	vnfState.VnfType = vnfCfg.VnfType
	vnfState.Group = vnfCfg.Group
	vnfState.VtepIP = vnfCfg.VtepIP
	vnfState.VnfLabels = make(map[string]string)
	for k, v := range vnfCfg.VnfLabels {
		vnfState.VnfLabels[k] = v
	}

	err = vnfState.Write()

	if err != nil {
		return err
	}

	return nil
}

// DeleteVnf deletes from etcd state
func DeleteVnf(vnfName string, tenantName string) error {

	log.Infof("Received Delete VNF %s on %s", vnfName, tenantName)

	// Get the state driver
	stateDriver, err := utils.GetStateDriver()
	if err != nil {
		return err
	}

	vnfState := &mastercfg.CfgVnfState{}
	vnfState.StateDriver = stateDriver
	vnfState.ID = GetVnfID(tenantName, vnfName)

	err = vnfState.Read(vnfState.ID)
	if err != nil {
		log.Errorf("Error reading vnf config for %s in tenant %s", vnfName, tenantName)
		return err
	}

	err = vnfState.Clear()
	if err != nil {
		log.Errorf("Error deleting VNF config for vnf %s in tenant %s", vnfName, tenantName)
		return err
	}

	return nil

}

// GetVnfID returns VNF ID for state lookup
func GetVnfID(tenantName string, vnfName string) string {
	return tenantName + ":" + vnfName
}
