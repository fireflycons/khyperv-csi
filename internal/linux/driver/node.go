//go:build linux

/*
Copyright 2022 DigitalOcean

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

// Modified by firefycons, based on https://github.com/digitalocean/csi-digitalocean/blob/master/driver/node.go
// Tweaks mainly for Go v1.25 otherwise largely unmodified

package driver

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	mountutil "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	diskIDPath = "/dev/disk/by-id"

	volumeModeBlock      = "block"
	volumeModeFilesystem = "filesystem"
)

var (
	// This annotation is added to a PV to indicate that the volume should be
	// not formatted. Useful for cases if the user wants to reuse an existing
	// volume. We support using either the legacy driver name
	// (io.fireflycons.csi.hyperv) or the modern driver name
	// (hyperv.csi.fireflycons.io).
	annsNoFormatVolume = []string{
		DefaultDriverName + "/noformat",
		DefaultDriverNameRDN + "/noformat",
	}
)

// NodeStageVolume mounts the volume to a staging path on the node. This is
// called by the CO before NodePublishVolume and is used to temporary mount the
// volume to a staging path. Once mounted, NodePublishVolume will make sure to
// mount it to the appropriate path
func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) { //nolint:gocyclo // cyclomatic complexity is ok here
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Staging Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume Capability must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": req.StagingTargetPath,
		"method":              "node_stage_volume",
	})
	log.Info("node stage volume called")

	volumeName := ""
	if volName, ok := req.GetPublishContext()[d.publishInfoVolumeName]; !ok {
		return nil, status.Error(codes.InvalidArgument, "Could not find the volume by name")
	} else {
		volumeName = volName
	}

	// If it is a block volume, we do nothing for stage volume
	// because we bind mount the absolute device path to a file
	if _, ok := req.GetVolumeCapability().GetAccessType().(*csi.VolumeCapability_Block); ok {
		return &csi.NodeStageVolumeResponse{}, nil
	}

	source, err := hypervDiskByID(req.VolumeId)
	if err != nil {
		return nil, fmt.Errorf("cannot determine udev disk ID: %w", err)
	}

	target := req.StagingTargetPath

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	fsType := fstypeExt4
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	log = d.log.WithFields(logrus.Fields{
		"volume_mode":     volumeModeFilesystem,
		"volume_name":     volumeName,
		"volume_context":  req.VolumeContext,
		"publish_context": req.PublishContext,
		"source":          source,
		"fs_type":         fsType,
		"mount_options":   options,
	})

	log.Info("checking if the disk needs formatting")

	var noFormat bool
	for _, ann := range annsNoFormatVolume {
		_, noFormat = req.VolumeContext[ann]
		if noFormat {
			break
		}
	}
	if noFormat {
		log.Info("skipping formatting the source device")
	} else {
		if d.validateAttachment {
			//nolint:govet // intentional redeclaration of err
			if err := d.mounter.IsAttached(source); err != nil {
				return nil, fmt.Errorf("error retrieving the attachement status %q: %w", source, err)
			}
		}

		//nolint:govet // intentional redeclaration of err
		formatted, err := d.mounter.IsFormatted(source)
		if err != nil {
			return nil, err
		}

		if !formatted {
			log.Info("formatting the volume for staging")
			if err := d.mounter.Format(source, fsType); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			log.Info("source device is already formatted")
		}
	}

	log.Info("mounting the volume for staging")

	mounted, err := d.mounter.IsMounted(target)
	if err != nil {
		return nil, err
	}

	if !mounted {
		if err := d.mounter.Mount(source, target, fsType, options...); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		log.Info("source device is already mounted to the target path")
	}

	if _, err := os.Stat(source); err == nil {
		r := mountutil.NewResizeFs(utilexec.New())
		needResize, err := r.NeedResize(source, target)

		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not determine if volume %q need to be resized: %v", req.VolumeId, err)
		}

		if needResize {
			//nolint:mnd // Enum not defined for this
			klog.V(4).Infof("NodeStageVolume: Resizing volume %q created from a snapshot/volume", req.VolumeId)
			if _, err := r.Resize(source, target); err != nil {
				return nil, status.Errorf(codes.Internal, "Could not resize volume %q:  %v", req.VolumeId, err)
			}
		}
	}

	log.Info("formatting and mounting stage volume is finished")
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unstages the volume from the staging path
func (d *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Staging Target Path must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": req.StagingTargetPath,
		"method":              "node_unstage_volume",
	})
	log.Info("node unstage volume called")

	mounted, err := d.mounter.IsMounted(req.StagingTargetPath)
	if err != nil {
		return nil, err
	}

	if mounted {
		log.Info("unmounting the staging target path")
		err := d.mounter.Unmount(req.StagingTargetPath)
		if err != nil {
			return nil, err
		}
	} else {
		log.Info("staging target path is already unmounted")
	}

	log.Info("unmounting stage volume is finished")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Staging Target Path must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume Capability must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": req.StagingTargetPath,
		"target_path":         req.TargetPath,
		"method":              "node_publish_volume",
	})
	log.Info("node publish volume called")

	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}

	var err error
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		err = d.nodePublishVolumeForBlock(req, options, log)
	case *csi.VolumeCapability_Mount:
		err = d.nodePublishVolumeForFileSystem(req, options, log)
	default:
		return nil, status.Error(codes.InvalidArgument, "Unknown access type")
	}

	if err != nil {
		return nil, err
	}

	log.Info("bind mounting the volume is finished")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Volume ID must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Target Path must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":   req.VolumeId,
		"target_path": req.TargetPath,
		"method":      "node_unpublish_volume",
	})
	log.Info("node unpublish volume called")

	err := d.mounter.Unmount(req.TargetPath)
	if err != nil {
		return nil, err
	}

	log.Info("unmounting volume is finished")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscaps := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				},
			},
		},
		// TODO - Expand Volume support
		// {
		// 	Type: &csi.NodeServiceCapability_Rpc{
		// 		Rpc: &csi.NodeServiceCapability_RPC{
		// 			Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
		// 		},
		// 	},
		// },
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
				},
			},
		},
	}

	d.log.WithFields(logrus.Fields{
		"node_capabilities": nscaps,
		"method":            "node_get_capabilities",
	}).Info("node get capabilities called")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nscaps,
	}, nil
}

// NodeGetInfo returns the supported capabilities of the node server. This is used so the CO
// knows where to place the workload. The result of this function will be used
// by the CO in ControllerPublishVolume.
func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	d.log.WithField("method", "node_get_info").Info("node get info called")
	return &csi.NodeGetInfoResponse{
		NodeId:            d.hostID(),
		MaxVolumesPerNode: int64(d.volumeLimit), //nolint:gosec // conversions are OK here
	}, nil
}

// NodeGetVolumeStats returns the volume capacity statistics available for the
// the given volume.
func (d *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume ID must be provided")
	}

	volumePath := req.VolumePath
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume Path must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":   req.VolumeId,
		"volume_path": req.VolumePath,
		"method":      "node_get_volume_stats",
	})
	log.Info("node get volume stats called")

	mounted, err := d.mounter.IsMounted(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if volume path %q is mounted: %s", volumePath, err)
	}

	if !mounted {
		return nil, status.Errorf(codes.NotFound, "volume path %q is not mounted", volumePath)
	}

	isBlock, err := d.mounter.IsBlockDevice(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to determine if %q is block device: %s", volumePath, err)
	}

	stats, err := d.mounter.GetStatistics(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve capacity statistics for volume path %q: %s", volumePath, err)
	}

	// only can retrieve total capacity for a block device
	if isBlock {
		log.WithFields(logrus.Fields{
			"volume_mode": volumeModeBlock,
			"bytes_total": stats.totalBytes,
		}).Info("node capacity statistics retrieved")

		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Unit:  csi.VolumeUsage_BYTES,
					Total: stats.totalBytes,
				},
			},
		}, nil
	}

	log.WithFields(logrus.Fields{
		"volume_mode":      volumeModeFilesystem,
		"bytes_available":  stats.availableBytes,
		"bytes_total":      stats.totalBytes,
		"bytes_used":       stats.usedBytes,
		"inodes_available": stats.availableInodes,
		"inodes_total":     stats.totalInodes,
		"inodes_used":      stats.usedInodes,
	}).Info("node capacity statistics retrieved")

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: stats.availableBytes,
				Total:     stats.totalBytes,
				Used:      stats.usedBytes,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: stats.availableInodes,
				Total:     stats.totalInodes,
				Used:      stats.usedInodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

func (d *Driver) nodePublishVolumeForFileSystem(req *csi.NodePublishVolumeRequest, mountOptions []string, log *logrus.Entry) error {
	source := req.StagingTargetPath
	target := req.TargetPath

	mnt := req.VolumeCapability.GetMount()
	mountOptions = append(mountOptions, mnt.MountFlags...)

	fsType := fstypeExt4
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	mounted, err := d.mounter.IsMounted(target)
	if err != nil {
		return err
	}

	log = log.WithFields(logrus.Fields{
		"source_path":   source,
		"volume_mode":   volumeModeFilesystem,
		"fs_type":       fsType,
		"mount_options": mountOptions,
	})

	if !mounted {
		log.Info("mounting the volume")
		if err := d.mounter.Mount(source, target, fsType, mountOptions...); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		log.Info("volume is already mounted")
	}

	return nil
}

func (d *Driver) nodePublishVolumeForBlock(req *csi.NodePublishVolumeRequest, mountOptions []string, log *logrus.Entry) error {
	volumeName, ok := req.GetPublishContext()[d.publishInfoVolumeName]
	if !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Could not find the volume name from the publish context %q", d.publishInfoVolumeName))
	}

	source, err := hypervDiskByID(req.VolumeId)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to find device path for volume %s. %v", volumeName, err)
	}

	target := req.TargetPath

	mounted, err := d.mounter.IsMounted(target)
	if err != nil {
		return err
	}

	log = log.WithFields(logrus.Fields{
		"source_path":   source,
		"volume_mode":   volumeModeBlock,
		"mount_options": mountOptions,
	})

	if !mounted {
		log.Info("mounting the volume")
		if err := d.mounter.Mount(source, target, "", mountOptions...); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		log.Info("volume is already mounted")
	}

	return nil
}

// hypervDiskByID converts a Hyper-V DiskIdentifier UUID into the /dev/disk/by-id
// "scsi-3<wwn>" link that Linux creates for synthetic SCSI disks.
//
// WWN layout (16 bytes total):
//
//	[0:4]   = 0x60 0x02 0x24 0x80  (NAA=6 + Microsoft OUI)
//	[4:8]   = UUID.Data1 (reversed for little-endian)
//	[8:10]  = UUID.Data2 (reversed for little-endian)
//	[10:16] = last 6 bytes of UUID.Data4 (drop first 2 bytes)
func hypervDiskByID(guid string) (string, error) {

	const (
		networkAddressAuthority = "\x60"
		microsoftOUI            = "\x02\x24\x80"

		wwnLength = 16
	)

	u, err := uuid.Parse(guid)
	if err != nil {
		return "", err
	}
	b := u[:] // 16-byte RFC 4122 layout

	// Reverse first three fields to convert RFC4122 -> Windows GUID memory layout.
	reverse := func(bs []byte) { // helper to reverse bytes in place
		for i, j := 0, len(bs)-1; i < j; i, j = i+1, j-1 {
			bs[i], bs[j] = bs[j], bs[i]
		}
	}
	reverse(b[0:4]) // Data1
	reverse(b[4:6]) // Data2
	reverse(b[6:8]) // Data3

	// Take Data4 (bytes[8:16]), but only the last 6 bytes.
	last6 := b[10:16]

	// Build WWN: 4-byte prefix + p1(4) + p2(2) + last6(6) => 16 bytes
	wwn := make([]byte, 0, wwnLength)
	wwn = append(wwn, []byte(networkAddressAuthority)...) // NAA type
	wwn = append(wwn, []byte(microsoftOUI)...)            // Microsoft OUI
	wwn = append(wwn, b[0:4]...)                          // Data1 (LE)
	wwn = append(wwn, b[4:6]...)                          // Data2 (LE)
	wwn = append(wwn, last6...)

	hexStr := hex.EncodeToString(wwn) // lowercase
	return filepath.Join(diskIDPath, fmt.Sprintf("scsi-3%s", hexStr)), nil
}
