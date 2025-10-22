//go:build linux

package driver

// Based on code from https://github.com/digitalocean/csi-digitalocean/blob/master/driver/controller.go

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/shared"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		restErr := &rest.Error{}
		if errors.As(err, &restErr) {
			log.WithError(err).Error("create volume failed")
			return nil, status.Errorf(restErr.Code, "create volume failed: %s", restErr.Message)
		}

		log.WithError(err).Error("create volume failed with unknown error")
		return nil, status.Errorf(codes.Internal, "create volume failed: %v", err)
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

// DeleteVolume deletes the given volume. The function is idempotent.
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
		restErr := &rest.Error{}
		if errors.As(err, &restErr) {
			log.WithError(err).Error("delete volume failed")
			return nil, status.Errorf(restErr.Code, "delete volume failed: %s", restErr.Message)
		}

		log.WithError(err).Error("delete volume failed with unknown error")
		return nil, status.Errorf(codes.Internal, "delete volume failed: %v", err)
	}

	log.Info("volume was deleted")
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume Volume ID must be provided")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ControllerPublishVolume Node ID must be provided")
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

	err := d.hypervClient.PublishVolume(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		restErr := &rest.Error{}
		if errors.As(err, &restErr) {
			log.WithError(err).Error("publish volume failed")
			return nil, status.Errorf(restErr.Code, "publish volume failed: %s", restErr.Message)
		}

		log.WithError(err).Error("publish volume failed with unknown error")
		return nil, status.Errorf(codes.Internal, "publish volume failed: %v", err)
	}

	log.Info("volume was published")
	// TODO Do we need to return the volume name as distinct from the ID?
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			d.publishInfoVolumeName: req.VolumeId,
		},
	}, nil
}

// ControllerUnpublishVolume deattaches the given volume from the node
func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ControllerUnpublishVolume Volume ID must be provided")
	}

	log := d.log.WithFields(logrus.Fields{
		"volume_id": req.VolumeId,
		"node_id":   req.NodeId,
		"method":    "controller_unpublish_volume",
	})
	log.Info("controller unpublish volume called")

	err := d.hypervClient.UnpublishVolume(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		restErr := &rest.Error{}
		if errors.As(err, &restErr) {
			log.WithError(err).Error("unpublish volume failed")
			return nil, status.Errorf(restErr.Code, "unpublish volume failed: %s", restErr.Message)
		}

		log.WithError(err).Error("unpublish volume failed with unknown error")
		return nil, status.Errorf(codes.Internal, "unpublish volume failed: %v", err)
	}

	log.Info("volume was unpublished")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// validateCapabilities validates the requested capabilities. It returns a list
// of violations which may be empty if no violatons were found.
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
func (d *Driver) extractStorage(capRange *csi.CapacityRange) (int64, error) {
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
		return 0, fmt.Errorf("limit (%v) can not be less than required (%v) size", shared.FormatBytes(limitBytes), shared.FormatBytes(requiredBytes))
	}

	if requiredSet && !limitSet && requiredBytes < constants.MinimumVolumeSizeInBytes {
		d.log.WithFields(logrus.Fields{
			"required_bytes":      shared.FormatBytes(requiredBytes),
			"minimum_volume_size": shared.FormatBytes(constants.MinimumVolumeSizeInBytes),
		}).Warn("requiredBytes is less than minimum volume size, setting requiredBytes default to minimumVolumeSizeBytes")
		return constants.MinimumVolumeSizeInBytes, nil
	}

	if limitSet && limitBytes < constants.MinimumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not be less than minimum supported volume size (%v)", shared.FormatBytes(limitBytes), shared.FormatBytes(constants.MinimumVolumeSizeInBytes))
	}

	if requiredSet && requiredBytes > constants.MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("required (%v) can not exceed maximum supported volume size (%v)", shared.FormatBytes(requiredBytes), shared.FormatBytes(constants.MaximumVolumeSizeInBytes))
	}

	if !requiredSet && limitSet && limitBytes > constants.MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not exceed maximum supported volume size (%v)", shared.FormatBytes(limitBytes), shared.FormatBytes(constants.MaximumVolumeSizeInBytes))
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
