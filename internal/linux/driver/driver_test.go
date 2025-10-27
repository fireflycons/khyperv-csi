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
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const (
	numVms = 100
)

var vms = make(map[int]string, numVms)

type idGenerator struct{}

func (*idGenerator) GenerateUniqueValidVolumeID() string {
	return uuid.New().String()
}

func (g *idGenerator) GenerateInvalidVolumeID() string {
	return g.GenerateUniqueValidVolumeID()
}

func (*idGenerator) GenerateUniqueValidNodeID() string {
	return uuid.New().String()
}

func (*idGenerator) GenerateInvalidNodeID() string {
	return "invalid id"
}

func TestDriverSuite(t *testing.T) {
	socket := "/tmp/csi.sock"
	endpoint := "unix://" + socket

	if err := os.Remove(socket); err != nil {
		require.ErrorIs(t, err, os.ErrNotExist, "failed to remove unix domain socket file %s", socket)
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

	l := logrus.New()

	// Comment these 2 lines to get log output
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	l.Out = devNull

	driver := &Driver{
		name:     DefaultDriverName,
		endpoint: endpoint,
		hostID: func() string {
			i := vmIdx % numVms
			vmIdx++
			return vms[i]
		},
		mounter:      fm,
		log:          l.WithField("test_enabed", true),
		hypervClient: client,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var eg errgroup.Group
	eg.Go(func() error {
		return driver.Run(ctx)
	})

	cfg := sanity.NewTestConfig()
	require.NoError(t, os.RemoveAll(cfg.TargetPath), "failed to delete target path %s", cfg.TargetPath)
	require.NoError(t, os.RemoveAll(cfg.StagingPath), "failed to delete staging path %s", cfg.StagingPath)

	cfg.Address = endpoint
	cfg.IDGen = &idGenerator{}
	cfg.IdempotentCount = 5
	cfg.TestNodeVolumeAttachLimit = true
	cfg.CheckPath = fm.checkMountPath
	cfg.TestVolumeSize = 50 * constants.MiB
	sanity.Test(t, cfg)

	cancel()
	require.NoError(t, eg.Wait(), "driver run failed: %s")
}
