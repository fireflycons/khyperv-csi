//go:build windows

package controller

import (
	"errors"
	"slices"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// ControllerServer implements the backend in Hyper-V land
// to the ControllerServer running in-cluster
type ControllerServer interface {
	HandleCreateVolume(*gin.Context)
	HandleGetVolume(*gin.Context)
	HandleDeleteVolume(*gin.Context)
	HandleListVolumes(*gin.Context)
	HandleGetCapacity(*gin.Context)
	HandlePublishVolume(*gin.Context)
	HandleUnpublishVolume(*gin.Context)
	HandleHealthCheck(*gin.Context)

	Logger() *logrus.Logger
	Close()
}

var _ ControllerServer = (*controllerServer)(nil)

type controllerServer struct {

	// Path to directory containing VHDs
	PVStore string

	// If this is non-nil then there was an error intitializing
	// the PV storage. All calls to the interface should return an error
	Err error

	runner powershell.Runner

	log *logrus.Logger
}

// NewController creates a new instance of the controller server
func NewController(logger *logrus.Logger, pvstore string) (*controllerServer, error) {

	runner, err := powershell.NewRunner(powershell.WithModules(constants.PowerShellModule))

	if err != nil {
		return nil, err
	}

	if pvstore == "" {
		// No user supplied store - let system choose.
		//nolint:govet // intentional redeclaration of err
		var err error
		pvstore, err = vhd.GetStorePath(runner)

		if err != nil {
			return nil, err
		}
	}

	logger.WithField("store", pvstore).Info("Selected PV store directory")
	return &controllerServer{
		PVStore: pvstore,
		log:     logger,
		runner:  runner,
		Err:     err,
	}, nil
}

// Close releases any resources associated with the controller server
func (s controllerServer) Close() {
	if s.runner != nil {
		s.runner.Exit()
	}
}

func (s controllerServer) Logger() *logrus.Logger {
	return s.log
}

// Log any error and convert to rest.Error for returning to the kube controller
func (*controllerServer) processError(err error, logEntry *logrus.Entry, message string, dontLogCodes ...codes.Code) *rest.Error {

	restErr := &rest.Error{}

	if !errors.As(err, &restErr) {
		restErr = &rest.Error{
			Code:    codes.Internal,
			Message: err.Error(),
		}
	}

	if !slices.Contains(dontLogCodes, restErr.Code) {
		logEntry.WithField("error", err.Error()).Error(message)
	}

	return restErr
}
