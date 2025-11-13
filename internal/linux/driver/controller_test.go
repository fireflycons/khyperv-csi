//go:build linux

package driver

import (
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestExtractStorage(t *testing.T) {

	logBuf := &strings.Builder{}
	logger := logrus.New()
	logger.Out = logBuf

	d := &Driver{
		log: logger.WithField("test", true),
	}

	requested := int64(constants.MiB * 25)

	actual, err := d.extractStorage(&csi.CapacityRange{
		RequiredBytes: requested,
		LimitBytes:    constants.MaximumVolumeSizeInBytes,
	})

	require.NoError(t, err)
	require.Equal(t, requested, actual)
}
