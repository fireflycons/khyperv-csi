//go:build linux

// The driver end to end functionality is tested by CSI sanity check
// package via Ginkgo
package driver

import (
	"context"
	"os"
	"testing"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/google/uuid"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	numVms = 100
)

var vms = make(map[int]string, numVms)

type idGenerator struct{}

func (g *idGenerator) GenerateUniqueValidVolumeID() string {
	return uuid.New().String()
}

func (g *idGenerator) GenerateInvalidVolumeID() string {
	return g.GenerateUniqueValidVolumeID()
}

func (g *idGenerator) GenerateUniqueValidNodeID() string {
	return vms[0]
}

func (g *idGenerator) GenerateInvalidNodeID() string {
	return uuid.New().String()
}

func TestDriverSuite(t *testing.T) {
	socket := "/tmp/csi.sock"
	endpoint := "unix://" + socket
	if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove unix domain socket file %s, error: %s", socket, err)
	}

	volumes := map[string]*models.GetVHDResponse{}

	for i := range numVms {
		vms[i] = uuid.New().String()
	}

	vmIdx := 0

	fm := &fakeMounter{
		mounted: map[string]string{},
	}

	client := &fakeClient{
		volumes: volumes,
		nodes:   vms,
	}

	driver := &Driver{
		name:     DefaultDriverName,
		endpoint: endpoint,
		hostID: func() string {
			i := vmIdx % numVms
			vmIdx++
			return vms[i]
		},
		mounter:      fm,
		log:          logrus.New().WithField("test_enabed", true),
		hypervClient: client,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var eg errgroup.Group
	eg.Go(func() error {
		return driver.Run(ctx)
	})

	cfg := sanity.NewTestConfig()
	if err := os.RemoveAll(cfg.TargetPath); err != nil {
		t.Fatalf("failed to delete target path %s: %s", cfg.TargetPath, err)
	}
	if err := os.RemoveAll(cfg.StagingPath); err != nil {
		t.Fatalf("failed to delete staging path %s: %s", cfg.StagingPath, err)
	}
	cfg.Address = endpoint
	cfg.IDGen = &idGenerator{}
	cfg.IdempotentCount = 5
	cfg.TestNodeVolumeAttachLimit = true
	cfg.CheckPath = fm.checkMountPath
	cfg.TestVolumeSize = 50 * constants.MiB
	sanity.Test(t, cfg)

	cancel()
	if err := eg.Wait(); err != nil {
		t.Errorf("driver run failed: %s", err)
	}
}
