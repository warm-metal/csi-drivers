/*
Copyright 2017 The Kubernetes Authors.

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

package csicommon

import (
	"fmt"
	"k8s.io/klog/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

type CSIDriver struct {
	name      string
	nodeID    string
	version   string
	cap       []*csi.ControllerServiceCapability
	vc        []*csi.VolumeCapability_AccessMode
	pluginCap []*csi.PluginCapability
}

// Creates a NewCSIDriver object. Assumes vendor version is equal to driver version &
// does not support optional driver plugin info manifest field. Refer to CSI spec for more details.
func NewCSIDriver(name string, v string, nodeID string) *CSIDriver {
	if name == "" {
		klog.Errorf("Driver name missing")
		return nil
	}

	if nodeID == "" {
		klog.Errorf("NodeID missing")
		return nil
	}
	// TODO version format and validation
	if len(v) == 0 {
		klog.Errorf("Version argument missing")
		return nil
	}

	driver := CSIDriver{
		name:    name,
		version: v,
		nodeID:  nodeID,
	}

	return &driver
}

func (d *CSIDriver) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range d.cap {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, fmt.Sprintf("%s", c))
}

func (d *CSIDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability

	for _, c := range cl {
		klog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}

	d.cap = csc

	return
}

func (d *CSIDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		klog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	d.vc = vca
	return vca
}

func (d *CSIDriver) GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	return d.vc
}

func (d *CSIDriver) AddPluginCapability(
	svc []csi.PluginCapability_Service_Type, volume []csi.PluginCapability_VolumeExpansion_Type,
) {
	pluginCap := make([]*csi.PluginCapability, 0, len(svc)+len(volume))
	for _, st := range svc {
		pluginCap = append(pluginCap, &csi.PluginCapability{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{Type: st},
			},
		})
	}

	for _, ve := range volume {
		pluginCap = append(pluginCap, &csi.PluginCapability{
			Type: &csi.PluginCapability_VolumeExpansion_{
				VolumeExpansion: &csi.PluginCapability_VolumeExpansion{Type: ve},
			},
		})
	}
	d.pluginCap = pluginCap
	return
}
