//go:build windows

package controller

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
)

// @BasePath		/
// @Summary		Create a new VHD
// @Param			X-Api-Key	header	string	true	"API Key"
// @Param			size		query	int		true	"Volume size"
// @Param			name		path	string	true	"Volume name"
// @Schemes		http
// @Description	Create a new VHD
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		201	{object}	rest.GetVolumeResponse
// @Failure		400	{object}	rest.Error "Invalid arguments"
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		409	{object}	rest.Error
// @Failure		500	{object}	rest.Error
// @Router			/volume/{name} [post]
func (s *controllerServer) HandleCreateVolume(ctx *gin.Context) {

	name := ctx.Param("name")

	if name == "" {
		abortInvalidArgument(ctx, "missing volume name")
		return
	}

	size := ctx.Query("size")

	if size == "" {
		abortInvalidArgument(ctx, "missing volume size")
		return
	}

	sizeBytes, err := strconv.ParseInt(size, 10, 64)

	if err != nil {
		abortArgumentError(ctx, fmt.Errorf("invalid volume size: %w", err))
		return
	}

	if sizeBytes < 0 {
		abortInvalidArgument(ctx, "volume size cannot be negative")
		return
	}

	resp, err := s.CreateVolume(name, sizeBytes)
	processResponse(ctx, resp, http.StatusCreated, err)
}

// @BasePath		/
// @Summary		Get an existing VHD
// @Param			X-Api-Key	header	string	true	"API Key"
// @Param			name		path	string	true	"Volume name or ID"
// @Schemes		http
// @Description	Get an existing VHD
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		201	{object}	rest.GetVolumeResponse
// @Failure		400	{object}	rest.Error "Invalid arguments"
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		404	{object}	rest.Error "Not found"
// @Failure		409	{object}	rest.Error
// @Failure		500	{object}	rest.Error
// @Router			/volume/{name} [get]
func (s *controllerServer) HandleGetVolume(ctx *gin.Context) {

	name := ctx.Param("name")

	if name == "" {
		abortInvalidArgument(ctx, "missing volume name")
		return
	}

	resp, err := s.GetVolume(name)
	processResponse(ctx, resp, http.StatusOK, err)
}

// @BasePath		/
// @Summary		Delete a VHD
// @Param			X-Api-Key	header	string	true	"API Key"
// @Param			id			path	string	true	"Volume ID"
// @Schemes		http
// @Description	Delete a VHD
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200
// @Failure		400	{object}	rest.Error "Invalid arguments"
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		500	{object}	rest.Error
// @Router			/volume/{id} [delete]
func (s *controllerServer) HandleDeleteVolume(ctx *gin.Context) {

	volId := ctx.Param("id")

	err := s.DeleteVolume(volId)
	processResponse(ctx, nil, http.StatusOK, err)
}

// @BasePath		/
// @Summary		List volumes
// @Param			X-Api-Key	header	string	true	"API Key"
// @Param			maxentries	query	int		false	"Maximum entires to return"
// @Param			nexttoken	query	string	false	"Next token for pagination"
// @Schemes		http
// @Description	List volumes
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200	{object}	rest.ListVolumesResponse
// @Failure		400	{object}	rest.Error "Invalid arguments"
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		500	{object}	rest.Error
// @Router			/volumes [get]
func (s *controllerServer) HandleListVolumes(ctx *gin.Context) {

	maxEntries := func() int {
		v := ctx.Query("maxentries")
		if v == "" {
			return 0
		}

		if m, err := strconv.Atoi(v); err != nil {
			ctx.JSON(http.StatusBadRequest, rest.Error{
				Code:    codes.InvalidArgument,
				Message: fmt.Errorf("invalid maxentries: %w", err).Error(),
			})
			return -1
		} else {
			return m
		}
	}()

	if maxEntries == -1 {
		return
	}

	me := func() int32 {
		v := maxEntries
		if v < 0 {
			v = -v
		}
		if v > math.MaxInt32 {
			return math.MaxInt32
		}

		//nolint:gosec // conversion is safe
		return int32(v)
	}()

	resp, err := s.ListVolumes(me, ctx.Query("nexttoken"))
	processResponse(ctx, resp, http.StatusOK, err)
}

// @BasePath		/
// @Summary		Get storage capacity
// @Param			X-Api-Key	header	string	true	"API Key"
// @Schemes		http
// @Description	Returns available capacity for new volumes, accounting for dynamic disk sizing
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200	{object}	rest.GetCapacityResponse
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		500	{object}	rest.Error
// @Router			/capacity [get]
func (s *controllerServer) HandleGetCapacity(ctx *gin.Context) {

	resp, err := s.GetCapacity()
	processResponse(ctx, resp, http.StatusOK, err)
}

// @BasePath		/
// @Summary		Publish Volume
// @Param			X-Api-Key	header	string	true	"API Key"
// @Schemes		http
// @Param			nodeid	path	string	true	"Node ID"
// @Param			volid	path	string	true	"Volume ID"
// @Description	Attaches a volume to a node
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		500	{object}	rest.Error
// @Router			/attachment/{nodeid}/volume/{volid} [post]
func (s *controllerServer) HandlePublishVolume(ctx *gin.Context) {

	err := s.PublishVolume(ctx.Param("volid"), ctx.Param("nodeid"))
	processResponse(ctx, nil, http.StatusOK, err)
}

// @BasePath		/
// @Summary		Unpublish Volume
// @Param			X-Api-Key	header	string	true	"API Key"
// @Schemes		http
// @Param			nodeid	path	string	true	"Node ID"
// @Param			volid	path	string	true	"Volume ID"
// @Description	Detaches a volume from a node
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200
// @Failure		403	{object}	rest.Error "Access denied"
// @Failure		500	{object}	rest.Error
// @Router			/attachment/{nodeid}/volume/{volid} [delete]
func (s *controllerServer) HandleUnpublishVolume(ctx *gin.Context) {

	err := s.UnpublishVolume(ctx.Param("volid"), ctx.Param("nodeid"))
	processResponse(ctx, nil, http.StatusOK, err)
}

// @BasePath		/
// @Summary		Check Health
// @Schemes		http
// @Description	Checks the health of the service
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200 {object}	rest.HealthyResponse
// @Failure		500	{object}	rest.Error
// @Router			/healthz [get]
func (s *controllerServer) HandleHealthCheck(ctx *gin.Context) {
	v := s.runner.Version()

	if v.Major < 0 {
		ctx.JSON(http.StatusInternalServerError, rest.Error{
			Code:    codes.Internal,
			Message: "Hyper-V backend unhealthy. Refer to eventlog on Hyper-V server",
		})
		return
	}

	ctx.JSON(http.StatusOK, rest.HealthyResponse{
		Status: "ok",
	})
}

// @BasePath		/
// @Summary		List Virtual Machines
// @Schemes		http
// @Description	Lists all VMs on the HYper-V server
// @Tags			Controller
// @Accept			json
// @Produce		json
// @Success		200 {object}	rest.ListVMResponse
// @Failure		500	{object}	rest.Error
// @Router			/vms [get]
func (s *controllerServer) HandleListVMs(ctx *gin.Context) {

	vms, err := s.ListVms()
	processResponse(ctx, vms, http.StatusOK, err)
}

func processResponse(ctx *gin.Context, response any, okStatus int, err error) {

	if err != nil {
		ctx.JSON(errorToHttpStatus(err), err)
		return
	}

	if response != nil {
		ctx.JSON(okStatus, response)
	} else {
		ctx.Status(okStatus)
	}
}

func abortInvalidArgument(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, &rest.Error{
		Code:    codes.InvalidArgument,
		Message: message,
	})
}

func abortArgumentError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, &rest.Error{
		Code:    codes.InvalidArgument,
		Message: err.Error(),
	})
}

// Map gRPC codes to appropriate HTTP status
var codeLookup = map[codes.Code]int{
	codes.OK:                 200,
	codes.Canceled:           500,
	codes.Unknown:            500,
	codes.InvalidArgument:    400,
	codes.DeadlineExceeded:   504,
	codes.NotFound:           404,
	codes.AlreadyExists:      409,
	codes.PermissionDenied:   403,
	codes.ResourceExhausted:  500,
	codes.FailedPrecondition: 412,
	codes.Aborted:            500,
	codes.OutOfRange:         400,
	codes.Unimplemented:      501,
	codes.Internal:           500,
	codes.Unavailable:        503,
	codes.DataLoss:           500,
	codes.Unauthenticated:    401,
}

func errorToHttpStatus(err error) int {

	restErr := &rest.Error{}

	if !errors.As(err, &restErr) {
		return http.StatusInternalServerError
	}

	if code, ok := codeLookup[restErr.Code]; ok {
		return code
	}

	return http.StatusInternalServerError
}
