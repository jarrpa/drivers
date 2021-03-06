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

package flexadapter

import (
	"os"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"

	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type flexAdapter struct {
	driver *csicommon.CSIDriver

	flexDriver *flexVolumeDriver

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

var (
	adapter *flexAdapter
	runOnce sync.Once
	version = csi.Version{
		Minor: 1,
	}
)

func GetSupportedVersions() []*csi.Version {
	return []*csi.Version{&version}
}

func GetFlexAdapter() *flexAdapter {
	runOnce.Do(func() {
		adapter = &flexAdapter{}
	})
	return adapter
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

func (f *flexAdapter) Run(driverName, driverPath, nodeID, endpoint string) {
	var err error

	glog.Infof("Driver: %v version: %v", driverName, GetVersionString(&version))

	// Create flex volume driver
	adapter.flexDriver, err = NewFlexVolumeDriver(driverName, driverPath)
	if err != nil {
		glog.Errorf("Failed to initialize flex volume driver, error: %v", err.Error())
		os.Exit(1)
	}

	// Initialize default library driver
	adapter.driver = csicommon.NewCSIDriver(driverName, &version, GetSupportedVersions(), nodeID)
	if adapter.flexDriver.capabilities.Attach {
		adapter.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME})
	}
	adapter.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	// Create GRPC servers
	f.ids = NewIdentityServer(adapter.driver)
	f.ns = NewNodeServer(adapter.driver)
	f.cs = NewControllerServer(adapter.driver)

	csicommon.Serve(endpoint, f.ids, f.cs, f.ns)
}
