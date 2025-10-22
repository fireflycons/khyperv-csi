//go:build linux

package driver

import (
	"context"
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
	"github.com/fireflycons/hypervcsi/internal/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	// DefaultDriverName defines the name that is used in Kubernetes and the CSI
	// system for the canonical, official name of this plugin
	DefaultDriverName = "hyperv.csi.fireflycons.io"

	defaultMaxVolumesPerNode = 256
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

	name string
	// publishInfoVolumeName is used to pass the volume name from
	// `ControllerPublishVolume` to `NodeStageVolume or `NodePublishVolume`
	publishInfoVolumeName string

	// unix socket endpoint
	endpoint               string
	debugAddr              string
	hostID                 func() string
	isController           bool
	defaultVolumesPageSize uint
	validateAttachment     bool

	srv *grpc.Server

	// To provide health check endpoint for readiness probes
	httpSrv *http.Server
	log     *logrus.Logger

	hypervClient hyperv.Client

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
	apiKey                 string
}

// NewDriver returns a CSI plugin that contains the necessary gRPC
// interfaces to interact with Kubernetes over unix domain sockets for
// managing Hyper-V Block Storage
func NewDriver(p NewDriverParams) (*Driver, error) {
	driverName := p.DriverName
	if driverName == "" {
		driverName = DefaultDriverName
	}

	log := logging.New()

	client, err := hyperv.NewClient(p.URL, &http.Client{}, p.apiKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create Hyper-V client: %w", err)
	}

	return &Driver{
		name:                   driverName,
		publishInfoVolumeName:  driverName + "/volume-name",
		endpoint:               p.Endpoint,
		debugAddr:              p.DebugAddr,
		defaultVolumesPageSize: p.DefaultVolumesPageSize,
		hypervClient:           client,
		log:                    log,
	}, nil
}

// Run starts the CSI plugin by communication over the given endpoint
func (d *Driver) Run(ctx context.Context) error {
	u, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("unable to parse address: %q", err)
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
	if err := os.Remove(grpcAddr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unix domain socket file %s, error: %s", grpcAddr, err)
	}

	grpcListener, err := net.Listen(u.Scheme, grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
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
				// TODO - Call backend for health check
				//
				// err := d.healthChecker.Check(r.Context())
				// if err != nil {
				// 	d.log.WithError(err).Error("executing health check")
				// 	http.Error(w, err.Error(), http.StatusInternalServerError)
				// 	return
				// }
				w.WriteHeader(http.StatusOK)
			})
			d.httpSrv = &http.Server{
				Addr:    d.debugAddr,
				Handler: mux,
			}
		}
	}

	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	csi.RegisterIdentityServer(d.srv, d)
	// csi.RegisterControllerServer(d.srv, d)
	// csi.RegisterNodeServer(d.srv, d)

	d.ready = true // we're now ready to go!
	d.log.WithFields(logrus.Fields{
		"grpc_addr": grpcAddr,
		"http_addr": d.debugAddr,
	}).Info("starting server")

	var eg errgroup.Group
	if d.httpSrv != nil {
		eg.Go(func() error {
			<-ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
