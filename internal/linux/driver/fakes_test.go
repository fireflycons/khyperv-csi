//go:build linux

// The driver end to end functionality is tested by CSI sanity check
// package via Ginkgo
package driver

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/hyperv"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/google/uuid"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
	"google.golang.org/grpc/codes"
	"k8s.io/mount-utils"
)

type fakeClient struct {
	volumes         map[string]*models.GetVHDResponse
	nodes           map[int]string
	createVolumeErr *rest.Error
	listVolumesErr  *rest.Error
}

var _ hyperv.Client = (*fakeClient)(nil)

func (f *fakeClient) HealthCheck(context.Context) (*rest.HealthyResponse, error) {
	return &rest.HealthyResponse{
		Status: "ok",
	}, nil
}

func (f *fakeClient) ListVolumes(_ context.Context, maxEntries int, nextToken string) (*rest.ListVolumesResponse, error) {

	if f.listVolumesErr != nil {
		return nil, f.listVolumesErr
	}

	if nextToken != "" {
		// Validate nextToken. The powershell expects an integer
		if _, err := strconv.Atoi(nextToken); err != nil {
			return nil, &rest.Error{
				Code:    codes.Aborted,
				Message: "Invalid starting token",
			}
		}
	}
	var volumes []*models.GetVHDResponse

	for _, vol := range f.volumes {
		volumes = append(volumes, vol)
	}

	if maxEntries > 0 {
		chunkSize := maxEntries
		nextToken := ""
		if len(volumes) < maxEntries {
			chunkSize = len(volumes)
			nextToken = strconv.Itoa(maxEntries)
		}

		return &rest.ListVolumesResponse{
			Volumes:   volumes[:chunkSize],
			NextToken: nextToken,
		}, nil
	}

	return &rest.ListVolumesResponse{
		Volumes:   volumes,
		NextToken: "",
	}, nil
}

func (f *fakeClient) CreateVolume(_ context.Context, name string, sizeBytes int64) (*rest.GetVolumeResponse, error) {

	if f.createVolumeErr != nil {
		return nil, f.createVolumeErr
	}

	// Idempotency check
	// Since CreateVolume doesn't know the ID before the backend is called
	// this check needs to be done here.
	for _, v := range f.volumes {
		if v.Name == name {
			if v.Size == sizeBytes {
				return volumeResponseFromVHD(v), nil
			}
			// Same name, different size is an error
			return nil, &rest.Error{
				Code:    codes.AlreadyExists,
				Message: "volume exists with different properties",
			}
		}
	}

	newId := uuid.NewString()
	path := fmt.Sprintf("c:\\disks\\%s;%s.vhdx", randString(8), newId)

	vol := &models.GetVHDResponse{
		Name:           name,
		Size:           sizeBytes,
		DiskIdentifier: newId,
		Path:           path,
	}

	f.volumes[newId] = vol

	return volumeResponseFromVHD(vol), nil
}

func volumeResponseFromVHD(vol *models.GetVHDResponse) *rest.GetVolumeResponse {
	return &rest.GetVolumeResponse{
		Name: vol.Name,
		ID:   vol.DiskIdentifier,
		Size: vol.Size,
	}
}

func (f *fakeClient) DeleteVolume(_ context.Context, volumeId string) error {
	delete(f.volumes, volumeId)
	return nil
}

func (f *fakeClient) GetVolume(_ context.Context, volumeId string) (*rest.GetVolumeResponse, error) {

	if v, ok := f.volumes[volumeId]; ok {
		return &rest.GetVolumeResponse{
			Name: v.Name,
			ID:   v.DiskIdentifier,
			Size: v.Size,
		}, nil
	} else {
		return nil, &rest.Error{
			Code:    codes.NotFound,
			Message: fmt.Sprintf("volume %s not found", volumeId),
		}
	}
}

func (f *fakeClient) GetCapacity(_ context.Context) (*rest.GetCapacityResponse, error) {
	return &rest.GetCapacityResponse{
		AvailableCapacity: constants.TiB,
		MinimumVolumeSize: constants.MinimumVolumeSizeInBytes,
	}, nil
}

func (f *fakeClient) PublishVolume(_ context.Context, volumeId, nodeId string) error {

	v, ok := f.volumes[volumeId]

	if !ok {
		// TODO - Check return of Add-VMHardDiskDrive when disk not found
		return &rest.Error{
			Code:    codes.NotFound,
			Message: fmt.Sprintf("volume %s not found", volumeId),
		}
	}

	// Idempotency check
	// TODO - In the controller, not here
	if v.Host != nil && *v.Host == nodeId {
		return nil
	}

	if v.Host == nil {
		v.Host = &nodeId
		return nil
	}

	return &rest.Error{
		Code:    codes.FailedPrecondition,
		Message: "The disk is already connected",
	}
}

func (f *fakeClient) UnpublishVolume(ctx context.Context, volumeId, nodeId string) error {

	if v, ok := f.volumes[volumeId]; ok {
		v.Host = nil
	}

	return nil
}

func (f *fakeClient) GetVm(_ context.Context, nodeId string) (*rest.GetVMResponse, error) {

	for _, n := range f.nodes {
		if strings.EqualFold(n, nodeId) {
			return &rest.GetVMResponse{
				ID: nodeId,
			}, nil
		}
	}

	return nil, &rest.Error{
		Code:    codes.NotFound,
		Message: "Node not found",
	}
}

func (f *fakeClient) ListVms(_ context.Context) (*rest.ListVMResponse, error) {
	vms := make([]*rest.GetVMResponse, 0, len(f.nodes))

	for _, n := range f.nodes {
		vms = append(vms, &rest.GetVMResponse{ID: n})
	}

	return &rest.ListVMResponse{
		VMs: vms,
	}, nil
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type fakeMounter struct {
	mounted map[string]string
}

var _ Mounter = (*fakeMounter)(nil)

func (f *fakeMounter) Format(source string, fsType string) error {
	return nil
}

func (f *fakeMounter) Mount(source string, target string, fsType string, options ...string) error {
	f.mounted[target] = source
	return nil
}

func (f *fakeMounter) Unmount(target string) error {
	delete(f.mounted, target)
	return nil
}

func (f *fakeMounter) GetDeviceName(_ mount.Interface, mountPath string) (string, error) {
	if _, ok := f.mounted[mountPath]; ok {
		return "/mnt/sda1", nil
	}

	return "", nil
}

func (f *fakeMounter) IsAttached(source string) error {
	return nil
}

func (f *fakeMounter) IsFormatted(source string) (bool, error) {
	return true, nil
}
func (f *fakeMounter) IsMounted(target string) (bool, error) {
	_, ok := f.mounted[target]
	return ok, nil
}

func (f *fakeMounter) checkMountPath(path string) (sanity.PathKind, error) {
	isMounted, err := f.IsMounted(path)
	if err != nil {
		return "", err
	}
	if isMounted {
		return sanity.PathIsDir, nil
	}
	return sanity.PathIsNotFound, nil
}

func (f *fakeMounter) GetStatistics(volumePath string) (volumeStatistics, error) {
	return volumeStatistics{
		availableBytes: 3 * constants.GiB,
		totalBytes:     10 * constants.GiB,
		usedBytes:      7 * constants.GiB,

		availableInodes: 3000,
		totalInodes:     10000,
		usedInodes:      7000,
	}, nil
}

func (f *fakeMounter) IsBlockDevice(volumePath string) (bool, error) {
	return false, nil
}
