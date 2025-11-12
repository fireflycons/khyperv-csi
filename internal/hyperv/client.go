// Package hyperv implements the client to
// the rest service running on the Hyper-V server
package hyperv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/sirupsen/logrus"
)

const (
	maxOperationWaitTime = 30 * time.Second
)

type Client interface {

	// CreateVolume creates a new VHD with the given name and size
	CreateVolume(ctx context.Context, name string, sizeBytes int64) (*rest.GetVolumeResponse, error)

	// DeleteVolume deletes a VHD with the given ID
	DeleteVolume(ctx context.Context, volumeId string) error

	// GetVolume retrieves a VHD with the given ID
	GetVolume(ctx context.Context, volumeId string) (*rest.GetVolumeResponse, error)

	// ListVolumes returns a list of provisioned VHDs
	ListVolumes(ctx context.Context, maxEntries int, nextToken string) (*rest.ListVolumesResponse, error)

	// GetCapacity returns the free space remaining for provisioning new VHDs
	GetCapacity(ctx context.Context) (*rest.GetCapacityResponse, error)

	// PublishVolume mounts a volume to a node
	PublishVolume(ctx context.Context, volumeId, nodeId string) error

	// UnpublishVolume dismounts a volume from a node
	UnpublishVolume(ctx context.Context, volumeId, nodeId string) error

	// ExpandVolume enlarges a volume to the given size
	ExpandVolume(ctx context.Context, volumeId string, size int64) (*rest.ExpandVolumeResponse, error)

	// ListVms returns a list of all VMs defined in the Hyper-V server
	ListVms(ctx context.Context) (*rest.ListVMResponse, error)

	// GetVm gets the VM with the given ID
	GetVm(ctx context.Context, nodeId string) (*rest.GetVMResponse, error)

	// HealthCheck performs a health check on the Hyper-V REST service
	HealthCheck(ctx context.Context) (*rest.HealthyResponse, error)
}

type noResult struct{}

type client struct {
	httpClient httpClient
	addr       *url.URL
	apiKey     string
	logger     *logrus.Entry
}

var _ Client = (*client)(nil)

func NewClient(baseURL string, httpClient httpClient, apiKey string, logger *logrus.Entry) (*client, error) {

	parsedURL, err := url.Parse(baseURL)

	if err != nil {
		return nil, fmt.Errorf("new hyperv client: cannot parse base URL: %w", err)
	}

	return &client{
		httpClient: httpClient,
		addr:       parsedURL,
		apiKey:     apiKey,
		logger:     logger,
	}, nil
}

var errNegativeValue = errors.New("argument value cannot be negative")

// CreateVolume creates a new VHD with the given name and size
func (c client) CreateVolume(ctx context.Context, name string, sizeBytes int64) (*rest.GetVolumeResponse, error) {

	if sizeBytes < 0 {
		return nil, errNegativeValue
	}

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + name + "/size/" + strconv.FormatInt(sizeBytes, 10),
	})

	return apiCall[*rest.GetVolumeResponse](ctx, c, "create volume", target, "POST")
}

// DeleteVolume deletes a VHD with the given ID
func (c client) DeleteVolume(ctx context.Context, volumeId string) error {

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + volumeId,
	})

	_, err := apiCall[*noResult](ctx, c, "delete volume", target, "DELETE")
	return err
}

func (c client) GetVolume(ctx context.Context, volumeId string) (*rest.GetVolumeResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + volumeId,
	})

	return apiCall[*rest.GetVolumeResponse](ctx, c, "get volume", target, "GET")
}

// ListVolumes returns a list of provisioned VHDs
func (c client) ListVolumes(ctx context.Context, maxEntries int, nextToken string) (*rest.ListVolumesResponse, error) {

	if maxEntries < 0 {
		return nil, errNegativeValue
	}

	target := c.addr.ResolveReference(&url.URL{
		Path: "volumes",
		RawQuery: url.Values{
			"maxentries": {strconv.FormatInt(int64(maxEntries), 10)},
			"nexttoken":  {nextToken},
		}.Encode(),
	})

	return apiCall[*rest.ListVolumesResponse](ctx, c, "list volumes", target, "GET")
}

// GetCapacity returns the free space remaining for provisioning new VHDs
func (c client) GetCapacity(ctx context.Context) (*rest.GetCapacityResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "capacity",
	})

	return apiCall[*rest.GetCapacityResponse](ctx, c, "get capacity", target, "GET")
}

// ListVms returns a list of all VMs defined in the Hyper-V server
func (c client) ListVms(ctx context.Context) (*rest.ListVMResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "vms",
	})

	return apiCall[*rest.ListVMResponse](ctx, c, "list vms", target, "GET")
}

// GetVm gets the VM with the given ID
func (c client) GetVm(ctx context.Context, nodeId string) (*rest.GetVMResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "vms",
		RawQuery: url.Values{
			"id": {nodeId},
		}.Encode(),
	})

	return apiCall[*rest.GetVMResponse](ctx, c, "get vm", target, "GET")

}

type publishOp int

const (
	publish publishOp = iota
	unpublish
)

// PublishVolume mounts a volume to a node
func (c client) PublishVolume(ctx context.Context, volumeId, nodeId string) error {

	return c.publisher(ctx, volumeId, nodeId, publish)
}

// UnpublishVolume dismounts a volume from a node
func (c client) UnpublishVolume(ctx context.Context, volumeId, nodeId string) error {

	return c.publisher(ctx, volumeId, nodeId, unpublish)
}

// ExpandVolume expands a volume to the given new size
func (c client) ExpandVolume(ctx context.Context, volumeId string, sizeBytes int64) (*rest.ExpandVolumeResponse, error) {

	if sizeBytes < 0 {
		return nil, errNegativeValue
	}

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + volumeId + "/size/" + strconv.FormatInt(sizeBytes, 10),
	})

	return apiCall[*rest.ExpandVolumeResponse](ctx, c, "create volume", target, "PUT")
}

// HealthCheck performs a basic check on the backend REST service
func (c client) HealthCheck(ctx context.Context) (*rest.HealthyResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "healthz",
	})

	return apiCall[*rest.HealthyResponse](ctx, c, "health check", target, "GET")
}

func (c client) publisher(ctx context.Context, volumeId, nodeId string, op publishOp) error {

	method, opName := func() (string, string) {
		if op == publish {
			return "PUT", "publish"
		}
		return "DELETE", "unpublish"
	}()

	target := c.addr.ResolveReference(&url.URL{
		Path: "attachment/" + nodeId + "/volume/" + volumeId,
	})

	_, err := apiCall[*noResult](ctx, c, opName+" volume", target, method)
	return err
}

// apiCall prepares and executes an API call to the Hyper-V REST service.
// It handles timeouts, request creation, and response parsing.
func apiCall[T *Q, Q any](ctx context.Context, c client, operation string, target *url.URL, method string) (T, error) {

	var requestCtx = ctx

	if ctx == context.Background() || ctx == context.TODO() {
		var cancel context.CancelFunc
		requestCtx, cancel = context.WithTimeout(ctx, maxOperationWaitTime)
		defer cancel()
	}

	request, err := http.NewRequestWithContext(requestCtx, method, target.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("%s: cannot create request: %w", operation, err)
	}

	request.Header.Set(constants.ApiKeyHeader, c.apiKey)

	if c.logger != nil {
		c.logger.WithFields(logrus.Fields{
			"curl":      requestToCurl(request),
			"operation": operation,
		}).Debug("")
	}

	httpResponse, err := c.httpClient.Do(request)

	if err != nil {
		return nil, fmt.Errorf("%s: error making request: %w", operation, err)
	}

	var bodyData []byte

	if httpResponse.Body != nil {

		bodyData, err = io.ReadAll(httpResponse.Body)
		_ = httpResponse.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("%s: error reading result: %w", operation, err)
		}
	}

	if httpResponse.StatusCode >= http.StatusBadRequest {

		errorObj := &rest.Error{}

		if err := json.Unmarshal(bodyData, errorObj); err != nil {
			return nil, fmt.Errorf("%s: error unmarshaling error response: %w", operation, err)
		}

		return nil, errorObj
	}

	var q Q
	apiResponse := &q

	if len(bodyData) > 0 {
		// A response is expected
		if err := json.Unmarshal(bodyData, apiResponse); err != nil {
			return nil, fmt.Errorf("%s: error unmarshaling response data: %w", operation, err)
		}
	}

	return apiResponse, nil
}

func requestToCurl(req *http.Request) string {

	if req == nil {
		return ""
	}

	apiKey := req.Header.Get(constants.ApiKeyHeader)

	return fmt.Sprintf(
		"curl -X %s -H '%s: %s' %s",
		func() string {
			if strings.TrimSpace(req.Method) == "" {
				return "GET"
			}
			return strings.ToUpper(req.Method)
		}(),
		constants.ApiKeyHeader,
		common.Redact(apiKey),
		req.URL.String(),
	)
}
