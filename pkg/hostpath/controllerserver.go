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

package hostpath

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	deviceID      = "deviceID"
	provisionRoot = "/tmp/"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func GetVersionString(ver *csi.Version) string {
	return fmt.Sprintf("%d.%d.%d", ver.Major, ver.Minor, ver.Patch)
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.Driver.ValidateRequest(req.Version, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}
	volumeId := uuid.NewUUID().String()
	path := provisionRoot + volumeId
	err := os.MkdirAll(path, 0777)
	if err != nil {
		glog.V(3).Infof("failed to create volume: %v", err)
		return nil, err
	}
	glog.V(4).Infof("create volume %s", path)
	return &csi.CreateVolumeResponse{
		VolumeInfo: &csi.VolumeInfo{
			Id: volumeId,
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.Driver.ValidateRequest(req.Version, csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid delete volume req: %v", req)
		return nil, err
	}
	volumeId := req.VolumeId
	glog.V(4).Infof("deleting volume %s", volumeId)
	path := provisionRoot + volumeId
	os.RemoveAll(path)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	for _, cap := range req.VolumeCapabilities {
		if cap.GetAccessMode().GetMode() != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{false, ""}, nil
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{true, ""}, nil
}
