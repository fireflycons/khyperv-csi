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
	"time"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
)

const (
	maxOperationWaitTime = 30 * time.Second
)

type Client interface {

	// CreateVolume creates a new VHD with the given name and size
	CreateVolume(ctx context.Context, name string, sizeBytes int64) (*rest.CreateVolumeResponse, error)

	// DeleteVolume deletes a VHD with the given ID
	DeleteVolume(ctx context.Context, volumeId string) error

	// ListVolumes returns a list of provisioned VHDs
	ListVolumes(ctx context.Context, maxEntries int, nextToken string) (*rest.ListVolumesResponse, error)

	// GetCapacity returns the free space remaining for provisioning new VHDs
	GetCapacity(ctx context.Context) (*rest.GetCapacityResponse, error)

	// PublishVolume mounts a volume to a node
	PublishVolume(ctx context.Context, volumeId, nodeId string) error

	// UnpublishVolume dismounts a volume from a node
	UnpublishVolume(ctx context.Context, volumeId, nodeId string) error
}

type noResult struct{}

type client struct {
	client httpClient
	addr   *url.URL
	apiKey string
}

func NewClient(baseURL string, httpClient httpClient, apiKey string) (*client, error) {

	parsedURL, err := url.Parse(baseURL)

	if err != nil {
		return nil, fmt.Errorf("new hyperv client: cannot parse base URL: %w", err)
	}

	return &client{
		client: httpClient,
		addr:   parsedURL,
		apiKey: apiKey,
	}, nil
}

var errNegativeValue = errors.New("argument value cannot be negative")

// CreateVolume creates a new VHD with the given name and size
func (c client) CreateVolume(ctx context.Context, name string, sizeBytes int64) (*rest.CreateVolumeResponse, error) {

	if sizeBytes < 0 {
		return nil, errNegativeValue
	}

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + name,
		RawQuery: url.Values{
			"size": {strconv.FormatInt(sizeBytes, 10)},
		}.Encode(),
	})

	reqCtx, cancel := context.WithTimeout(ctx, maxOperationWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", target.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("create volume: cannot create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	response := &rest.CreateVolumeResponse{}
	return executeRequest(c, "create volume", req, response)
}

// DeleteVolume deletes a VHD with the given ID
func (c client) DeleteVolume(ctx context.Context, volumeId string) error {

	target := c.addr.ResolveReference(&url.URL{
		Path: "volume/" + volumeId,
	})

	reqCtx, cancel := context.WithTimeout(ctx, maxOperationWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "DELETE", target.String(), http.NoBody)

	if err != nil {
		return fmt.Errorf("delete volume: cannot create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	_, err = executeRequest(c, "delete volume", req, &noResult{})
	return err
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

	reqCtx, cancel := context.WithTimeout(ctx, maxOperationWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", target.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("list volumes: cannot create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	response := &rest.ListVolumesResponse{}
	return executeRequest(c, "list volumes", req, response)
}

// GetCapacity returns the free space remaining for provisioning new VHDs
func (c client) GetCapacity(ctx context.Context) (*rest.GetCapacityResponse, error) {

	target := c.addr.ResolveReference(&url.URL{
		Path: "capacity",
	})

	reqCtx, cancel := context.WithTimeout(ctx, maxOperationWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", target.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("get capacity: cannot create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	response := &rest.GetCapacityResponse{}

	return executeRequest(c, "get capacity", req, response)
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

func (c client) publisher(ctx context.Context, volumeId, nodeId string, op publishOp) error {

	method, opName := func() (string, string) {
		if op == publish {
			return "POST", "publish"
		}
		return "DELETE", "unpublish"
	}()

	target := c.addr.ResolveReference(&url.URL{
		Path: "attachment/" + nodeId + "/volume/" + volumeId,
	})

	reqCtx, cancel := context.WithTimeout(ctx, maxOperationWaitTime)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, target.String(), http.NoBody)

	if err != nil {
		return fmt.Errorf("%s volume: cannot create request: %w", opName, err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	_, err = executeRequest(c, opName+" volume", req, &noResult{})
	return err

}

// executeRequest handles the HTTP plunmbing and associated errors, resturning any response.
func executeRequest[T *Q, Q any](c client, operation string, request *http.Request, response T) (T, error) {

	resp, err := c.client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("%s: error making request: %w", operation, err)
	}

	bodyData, err := io.ReadAll(resp.Body)

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("%s: error reading result: %w", operation, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {

		errorObj := &rest.Error{}

		if err := json.Unmarshal(bodyData, errorObj); err != nil {
			return nil, fmt.Errorf("%s: error unmarshaling error response: %w", operation, err)
		}

		return nil, errorObj
	}

	if len(bodyData) > 0 {
		// A response is expected
		if err := json.Unmarshal(bodyData, response); err != nil {
			return nil, fmt.Errorf("%s: error unmarshaling response data: %w", operation, err)
		}
	}

	return response, nil
}
