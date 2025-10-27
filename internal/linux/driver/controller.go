//go:build linux

package driver

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

// Modified by firefycons, based on https://github.com/digitalocean/csi-digitalocean/blob/master/driver/controller.go
// General structure and helper functions retained.
// Core logic adapted for Hyper-V.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	// We currently only support a single volume to be attached to a single node
	// in read/write mode. This corresponds to `accessModes.ReadWriteOnce` in a
	// PVC resource on Kubernetes
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

type (
	volumeIdentifier string
	nodeIdentifier   string
)

// CreateVolume creates a new volume from the given request. The function is idempotent.
func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Volume capabilities must be provided")
	}

	if violations := validateCapabilities(req.VolumeCapabilities); len(violations) > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume capabilities cannot be satisified: %s", strings.Join(violations, "; ")))
	}

	size, err := d.extractStorage(req.CapacityRange)
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	volumeName := req.Name

	log := d.log.WithFields(logrus.Fields{
		"volume_name":             volumeName,
		"storage_size_giga_bytes": size / constants.GiB,
		"method":                  "create_volume",
		"volume_capabilities":     req.VolumeCapabilities,
	})
	log.Info("create volume called")

	// Call the backend to create the volume.
	// If it already exists and is the same size, it will return success and the existing volume.
	// If it already exists and is a different size, it will return an error.
	// Else it will attempt to create the volume and return the status

	vol, err := d.hypervClient.CreateVolume(ctx, volumeName, size)

	if err != nil {
		return nil, processErrorReturn(err, log, "create volume")
	}

	resp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      vol.ID,
			CapacityBytes: vol.Size,
		},
	}

	log.WithField("response", resp).Info("volume created successfully")
	return resp, nil
}

// DeleteVolume deletes the given volume. The function is idempotent,
// thus an invalid volume ID means nothing other than "it was already deleted"
func (d *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id": req.VolumeId,
		"method":    "delete_volume",
	})
	log.Info("delete volume called")

	err := d.hypervClient.DeleteVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, processErrorReturn(err, log, "delete volume")
	}

	log.Info("volume was deleted")
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	if err := validateIds("ControllerPublishVolume", volumeIdentifier(req.VolumeId), nodeIdentifier(req.NodeId)); err != nil {
		return nil, err
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume Volume capability must be provided")
	}

	if req.Readonly {
		// TODO we should return codes.InvalidArgument, but the CSI
		// test fails, because according to the CSI Spec, this flag cannot be
		// changed on the same volume. However we don't use this flag at all,
		// as there are no `readonly` attachable volumes.
		return nil, status.Error(codes.AlreadyExists, "read only Volumes are not supported")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id": req.VolumeId,
		"node_id":   req.NodeId,
		"method":    "controller_publish_volume",
	})
	log.Info("controller publish volume called")

	// Verify the volume exists
	_, err := d.hypervClient.GetVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, processErrorReturn(err, log, "publish volume - volume does not exist")
	}

	// Verify the node exists
	if _, err := d.hypervClient.GetVm(ctx, req.NodeId); err != nil {
		return nil, processErrorReturn(err, log, "publish volume - node does not exist")
	}

	if err := d.hypervClient.PublishVolume(ctx, req.VolumeId, req.NodeId); err != nil {
		return nil, processErrorReturn(err, log, "publish volume")
	}

	log.Info("volume was published")

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			d.publishInfoVolumeName: req.VolumeId,
		},
	}, nil
}

// ControllerUnpublishVolume deattaches the given volume from the node
func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	if err := validateIds("ControllerUnpublishVolume", volumeIdentifier(req.VolumeId), nodeIdentifier(req.NodeId)); err != nil {
		return nil, err
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id": req.VolumeId,
		"node_id":   req.NodeId,
		"method":    "controller_unpublish_volume",
	})
	log.Info("controller unpublish volume called")

	err := d.hypervClient.UnpublishVolume(ctx, req.VolumeId, req.NodeId)

	if err != nil {
		return nil, processErrorReturn(err, log, "unpublish volume")
	}

	log.Info("volume was unpublished")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks whether the volume capabilities requested are supported.
func (d *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	if err := validateIds("ControllerUnpublishVolume", volumeIdentifier(req.VolumeId)); err != nil {
		return nil, err
	}

	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "ValidateVolumeCapabilities Volume Capabilities must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id":              req.VolumeId,
		"volume_capabilities":    req.VolumeCapabilities,
		"supported_capabilities": supportedAccessMode,
		"method":                 "validate_volume_capabilities",
	})
	log.Info("validate volume capabilities called")

	// check if volume exists before trying to validate it
	_, err := d.hypervClient.GetVolume(ctx, req.VolumeId)

	if err != nil {
		return nil, processErrorReturn(err, log, "get volume")
	}

	// Since we don't have topology constraints, then because it exists, it's valid
	resp := &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: supportedAccessMode,
				},
			},
		},
	}

	log.WithField("confirmed", resp.Confirmed).Info("supported capabilities")
	return resp, nil
}

func (d *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	maxEntries := req.MaxEntries
	if maxEntries == 0 && d.defaultVolumesPageSize > 0 {
		maxEntries = int32(d.defaultVolumesPageSize) //nolint:gosec // conversions are OK here
	}

	log := d.log.WithFields(logrus.Fields{
		"max_entries":           req.MaxEntries,
		"effective_max_entries": maxEntries,
		"req_starting_token":    req.StartingToken,
		"method":                "list_volumes",
	})
	log.Info("list volumes called")

	volumesResp, err := d.hypervClient.ListVolumes(ctx, int(maxEntries), req.StartingToken)

	if err != nil {
		return nil, processErrorReturn(err, log, "list volumes")
	}

	resp := &csi.ListVolumesResponse{
		NextToken: volumesResp.NextToken,
		Entries:   make([]*csi.ListVolumesResponse_Entry, 0, len(volumesResp.Volumes)),
	}

	for _, v := range volumesResp.Volumes {
		resp.Entries = append(resp.Entries, &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				VolumeId:      v.DiskIdentifier,
				CapacityBytes: v.Size,
			},
		})
	}

	log.WithField("num_volume_entries", len(resp.Entries)).Info("volumes listed")
	return resp, nil
}

// GetCapacity returns the capacity of the storage pool
func (d *Driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	log := d.log.WithFields(logrus.Fields{
		"params": req.Parameters,
		"method": "get_capacity",
	})
	log.Info("get capacity called")

	capResp, err := d.hypervClient.GetCapacity(ctx)

	if err != nil {
		return nil, processErrorReturn(err, log, "get capacity")
	}

	log.WithField("available_capacity_bytes", capResp.AvailableCapacity).Info("capacity retrieved")

	return &csi.GetCapacityResponse{
		AvailableCapacity: capResp.AvailableCapacity,
		MaximumVolumeSize: &wrapperspb.Int64Value{
			Value: constants.MaximumVolumeSizeInBytes,
		},
		MinimumVolumeSize: &wrapperspb.Int64Value{
			Value: capResp.MinimumVolumeSize,
		},
	}, nil
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (d *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var caps []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		// csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		// csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		// csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	d.log.WithFields(logrus.Fields{
		"response": resp,
		"method":   "controller_get_capabilities",
	}).Info("controller get capabilities called")
	return resp, nil
}

func processErrorReturn(err error, log *logrus.Entry, action string) error {

	restErr := &rest.Error{}
	if errors.As(err, &restErr) {
		log.WithError(err).Errorf("%s failed", action)
		return status.Errorf(restErr.Code, "%s failed: %s", action, restErr.Message)
	}

	log.WithError(err).Errorf("%s failed with unknown error", action)
	return status.Errorf(codes.Internal, "%s failed: %v", action, err)
}

// validateCapabilities validates the requested capabilities.
// It returns a list of violations which may be empty if no violations were found.
func validateCapabilities(caps []*csi.VolumeCapability) []string {
	violations := sets.NewString()
	for _, cap := range caps {
		if cap.GetAccessMode().GetMode() != supportedAccessMode.GetMode() {
			violations.Insert(fmt.Sprintf("unsupported access mode %s", cap.GetAccessMode().GetMode().String()))
		}

		accessType := cap.GetAccessType()
		switch accessType.(type) {
		case *csi.VolumeCapability_Block:
		case *csi.VolumeCapability_Mount:
		default:
			violations.Insert("unsupported access type")
		}
	}

	return violations.List()
}

// extractStorage extracts the storage size in bytes from the given capacity
// range. If the capacity range is not satisfied it returns the default volume
// size. If the capacity range is above supported sizes, it returns an
// error. If the capacity range is below supported size, it returns the minimum supported size
func (d *Driver) extractStorage(capRange *csi.CapacityRange) (int64, error) { //nolint:gocyclo // cyclomatic complexity is ok here
	if capRange == nil {
		return constants.DefaultVolumeSizeInBytes, nil
	}

	requiredBytes := capRange.GetRequiredBytes()
	requiredSet := 0 < requiredBytes
	limitBytes := capRange.GetLimitBytes()
	limitSet := 0 < limitBytes

	if !requiredSet && !limitSet {
		return constants.DefaultVolumeSizeInBytes, nil
	}

	if requiredSet && limitSet && limitBytes < requiredBytes {
		return 0, fmt.Errorf("limit (%v) can not be less than required (%v) size", common.FormatBytes(limitBytes), common.FormatBytes(requiredBytes))
	}

	if requiredSet && !limitSet && requiredBytes < constants.MinimumVolumeSizeInBytes {
		d.log.WithFields(logrus.Fields{
			"required_bytes":      common.FormatBytes(requiredBytes),
			"minimum_volume_size": common.FormatBytes(constants.MinimumVolumeSizeInBytes),
		}).Warn("requiredBytes is less than minimum volume size, setting requiredBytes default to minimumVolumeSizeBytes")
		return constants.MinimumVolumeSizeInBytes, nil
	}

	if limitSet && limitBytes < constants.MinimumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not be less than minimum supported volume size (%v)", common.FormatBytes(limitBytes), common.FormatBytes(constants.MinimumVolumeSizeInBytes))
	}

	if requiredSet && requiredBytes > constants.MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("required (%v) can not exceed maximum supported volume size (%v)", common.FormatBytes(requiredBytes), common.FormatBytes(constants.MaximumVolumeSizeInBytes))
	}

	if !requiredSet && limitSet && limitBytes > constants.MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not exceed maximum supported volume size (%v)", common.FormatBytes(limitBytes), common.FormatBytes(constants.MaximumVolumeSizeInBytes))
	}

	if requiredSet && limitSet && requiredBytes == limitBytes {
		return requiredBytes, nil
	}

	if requiredSet {
		return requiredBytes, nil
	}

	if limitSet {
		return limitBytes, nil
	}

	return constants.DefaultVolumeSizeInBytes, nil
}

func validateIds(op string, ids ...any) error {

	for _, id := range ids {
		var idType, idValue string

		switch v := id.(type) {
		case volumeIdentifier:
			idType = "volume"
			idValue = string(v)
		case nodeIdentifier:
			idType = "node"
			idValue = string(v)
		default:
			continue
		}

		if idValue == "" {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("%s %s ID must be provided", op, idType))
		}

		if !isValidId(idValue) {
			// For CSI compliance, an invalid ID format is the same as not found
			// since it woould indeed not be found.
			return status.Error(codes.NotFound, fmt.Sprintf("%s invalid %s ID", op, idType))
		}
	}

	return nil
}

// isValidId checks whether a given volume or node ID is in UUID format
func isValidId(id string) bool {

	_, err := uuid.Parse(id)
	return err == nil
}
