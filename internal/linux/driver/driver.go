//go:build linux

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

// Modified by firefycons, based on https://github.com/digitalocean/csi-digitalocean/blob/master/driver/driver.go

package driver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fireflycons/hypervcsi/internal/hyperv"
	"github.com/fireflycons/hypervcsi/internal/linux/kvp"
	"github.com/fireflycons/hypervcsi/internal/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	// DefaultDriverName defines the name that is used in Kubernetes and the CSI
	// system for the canonical, official name of this plugin
	DefaultDriverName    = "hyperv.csi.fireflycons.io"
	DefaultDriverNameRDN = "io.fireflycons.csi.hyperv"

	defaultMaxVolumesPerNode = 256

	defaultVolumesPageSize = 50
)

var (
	version = "0.0.1-debug"
)

// Driver implements the following CSI interfaces:
//
//	csi.IdentityServer
//	csi.ControllerServer
//	csi.NodeServer
type Driver struct {

	// Cover all unimplemented methods
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer

	// name is the name of the driver
	name string

	// vmName is the cached value retrieved from the KVP metadata service
	vmName string

	// vmId is the cached value retrieved from the KVP metadata service
	vmId string

	// publishInfoVolumeName is used to pass the volume name from
	// `ControllerPublishVolume` to `NodeStageVolume or `NodePublishVolume`
	publishInfoVolumeName string

	// unix socket endpoint
	endpoint string

	debugAddr              string
	hostID                 func() string
	isController           bool
	defaultVolumesPageSize uint
	validateAttachment     bool

	srv *grpc.Server

	metadata kvp.MetadataService

	// To provide health check endpoint for readiness probes
	httpSrv *http.Server
	log     *logrus.Entry

	hypervClient hyperv.Client

	mounter Mounter

	healthChecker *HealthChecker

	// ready defines whether the driver is ready to function. This value will
	// be used by the `Identity` service via the `Probe()` method.
	readyMu     sync.Mutex // protects ready
	ready       bool
	volumeLimit uint
}

// NewDriverParams defines the parameters that can be passed to NewDriver.
type NewDriverParams struct {
	Endpoint               string
	URL                    string
	DriverName             string
	DebugAddr              string
	DefaultVolumesPageSize uint
	ValidateAttachment     bool
	VolumeLimit            uint
	metadata               kvp.MetadataService
	apiKey                 string
}

// NewDriver returns a CSI plugin that contains the necessary gRPC
// interfaces to interact with Kubernetes over unix domain sockets for
// managing Hyper-V Block Storage
func NewDriver(p *NewDriverParams) (*Driver, error) {

	driverName := p.DriverName
	if driverName == "" {
		driverName = DefaultDriverName
	}

	if p.DefaultVolumesPageSize == 0 {
		p.DefaultVolumesPageSize = defaultVolumesPageSize
	}

	log := logging.New()

	md := p.metadata
	if md == nil {
		md = kvp.New()
	}

	if !md.IsPresent() {
		log.Error("Hyper-V KVP metadata service is not present; the driver cannot function")
		return nil, errors.New("hyper-v kvp metadata service is not present")
	}

	vmName, err := md.Find(kvp.VM_NAME_KEY)
	if err != nil {
		log.WithError(err).Error("cannot retrieve VM name from Hyper-V KVP metadata service")
		return nil, fmt.Errorf("cannot retrieve VM name from Hyper-V KVP metadata service: %w", err)
	}

	vmId, err := md.Find(kvp.VM_ID_KEY)
	if err != nil {
		log.WithError(err).Error("cannot retrieve VM ID from Hyper-V KVP metadata service")
		return nil, fmt.Errorf("cannot retrieve VM ID from Hyper-V KVP metadata service: %w", err)
	}

	logEntry := log.WithFields(logrus.Fields{
		"vm_name": vmName,
		"vm_id":   vmId,
	})

	hyperVClient, err := hyperv.NewClient(p.URL, &http.Client{}, p.apiKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create Hyper-V client: %w", err)
	}

	return &Driver{
		name:                   driverName,
		vmName:                 vmName,
		vmId:                   vmId,
		publishInfoVolumeName:  driverName + "/volume-name",
		endpoint:               p.Endpoint,
		debugAddr:              p.DebugAddr,
		defaultVolumesPageSize: defaultVolumesPageSize,
		hypervClient:           hyperVClient,
		log:                    logEntry,
		mounter:                newMounter(logEntry),
		metadata:               md,
		hostID: func() string {
			// This should not error because we already tested it during initialization
			id, _ := md.Find(kvp.VM_ID_KEY)
			return id
		},
		healthChecker: NewHealthChecker(&hvHealthChecker{client: hyperVClient}),
	}, nil
}

// Run starts the CSI plugin by communication over the given endpoint
func (d *Driver) Run(ctx context.Context) error {
	u, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("unable to parse address: %w", err)
	}

	grpcAddr := path.Join(u.Host, filepath.FromSlash(u.Path))
	if u.Host == "" {
		grpcAddr = filepath.FromSlash(u.Path)
	}

	// CSI plugins talk only over UNIX sockets currently
	if u.Scheme != "unix" {
		return fmt.Errorf("currently only unix domain sockets are supported, have: %s", u.Scheme)
	}
	// remove the socket if it's already there. This can happen if we
	// deploy a new version and the socket was created from the old running
	// plugin.
	d.log.WithField("socket", grpcAddr).Info("removing socket")

	//nolint:govet // intentional redeclaration of err
	if err := os.Remove(grpcAddr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unix domain socket file %s, error: %w", grpcAddr, err)
	}

	//nolint:noctx // no context is ok, for now
	grpcListener, err := net.Listen(u.Scheme, grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// log response errors for better observability
	errHandler := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			d.log.WithError(err).WithField("method", info.FullMethod).Error("method failed")
		}
		return resp, err
	}

	// warn the user, it'll not propagate to the user but at least we see if
	// something is wrong in the logs. Only check if the driver is running with
	// a token (i.e: controller)
	if d.isController {

		if d.debugAddr != "" {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

				err := d.healthChecker.Check(r.Context())
				if err != nil {
					d.log.WithError(err).Error("executing health check")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
			d.httpSrv = &http.Server{
				Addr:              d.debugAddr,
				Handler:           mux,
				ReadHeaderTimeout: time.Second * 2,
			}
		}
	}

	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	d.ready = true // we're now ready to go!
	d.log.WithFields(logrus.Fields{
		"grpc_addr": grpcAddr,
		"http_addr": d.debugAddr,
	}).Info("starting server")

	var eg errgroup.Group
	if d.httpSrv != nil {
		eg.Go(func() error {
			const shutdownTimeout = 10 * time.Second
			<-ctx.Done()
			//nolint:govet // intentional redeclaration of ctx. Previous usage is finished with.
			ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			return d.httpSrv.Shutdown(ctx)
		})
		eg.Go(func() error {
			err := d.httpSrv.ListenAndServe()
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		})
	}
	eg.Go(func() error {
		go func() {
			<-ctx.Done()
			d.log.Info("server stopped")
			d.readyMu.Lock()
			d.ready = false
			d.readyMu.Unlock()
			d.srv.GracefulStop()
		}()
		return d.srv.Serve(grpcListener)
	})

	return eg.Wait()
}
