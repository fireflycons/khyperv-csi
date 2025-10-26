//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/fireflycons/hypervcsi/internal/logging"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/controller"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/swaggerui"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

// hyperVService represents the Windows service for
// serving Hyper-V virtual disk interface via rest.
//
// Implements Handler interface in golang.org/x/sys/windows/svc
type hyperVService struct {
	controller controller.ControllerServer
}

func (s *hyperVService) Logger() *logrus.Logger {
	return s.controller.Logger()
}

func (s *hyperVService) Execute(_ []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	wg.Go(func() {
		s.serve(ctx, changes, cancel)
	})

	defer func() {
		wg.Wait()
		s.controller.Close()
	}()

	for c := range requests {
		switch c.Cmd {

		case svc.Interrogate:

			changes <- c.CurrentStatus
			// // Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
			// time.Sleep(100 * time.Millisecond)
			// changes <- c.CurrentStatus

		case svc.Stop, svc.Shutdown:

			cancel()
			changes <- svc.Status{State: svc.StopPending}
			return ssec, errno

		default:

			s.Logger().Error(fmt.Sprintf("unexpected control request #%d", c))
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return ssec, errno
}

func runService(name string, isDebug bool) {
	var (
		logger *logrus.Logger
		err    error
	)

	if isDebug {
		logger = logging.NewDebug()
	} else {
		logger = logging.New()
	}

	logger.Info(fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}

	cntrl, err := controller.NewController(logger, pvDirectoryFlag)

	if err != nil {
		logger.Error(fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	err = run(
		name,
		&hyperVService{
			controller: cntrl,
		},
	)
	if err != nil {
		logger.Error(fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s service stopped", name))
}

func (s *hyperVService) serve(ctx context.Context, changes chan<- svc.Status, cancel context.CancelFunc) {

	const serverShutdownGracePeriod = 5 * time.Second

	httpServer := s.runServer(changes, cancel)

	// Listen for the interrupt signal.
	// This call blocks until signal is raised.
	<-ctx.Done()

	// Log that we are shutting down
	s.Logger().Info(messages.SERVER_STOPPING)

	// This context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownGracePeriod)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		// CTRL-C pressed twice or context timed out.
		s.Logger().WithError(err).Error(messages.SERVER_FORCED_SHUTDOWN)
	}

	s.Logger().Info(messages.SERVER_EXIT)
}

func (s *hyperVService) runServer(changes chan<- svc.Status, cancel context.CancelFunc) *http.Server {

	router := gin.New()
	router.Use(apiKeyMiddleware(s.Logger(), apiKeyFlag), gin.Recovery())

	// Add Swagger
	swaggerui.SwaggerInfo.BasePath = "/"

	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", portFlag)),
		ginSwagger.DefaultModelsExpandDepth(-1))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/volume/:name", s.controller.HandleGetVolume)
	router.POST("/volume/:name", s.controller.HandleCreateVolume)
	router.DELETE("/volume/:id", s.controller.HandleDeleteVolume)
	router.GET("/volumes", s.controller.HandleListVolumes)
	router.GET("/capacity", s.controller.HandleGetCapacity)
	router.POST("/attachment/:nodeid/volume/:volid", s.controller.HandlePublishVolume)
	router.DELETE("/attachment/:nodeid/volume/:volid", s.controller.HandleUnpublishVolume)
	router.GET("/healthz", s.controller.HandleHealthCheck)
	router.GET("/vms", s.controller.HandleListVMs)
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/swagger/index.html")
	})

	// Shutdown on context cancel

	useSSL := certFlag != "" && keyFlag != ""

	// Setup HTTP server
	const readHeaderTimeout = 5 * time.Second

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", portFlag),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Run server as background task
	go func() {
		s.Logger().
			WithField("port", portFlag).
			WithField("ssl", useSSL).
			Info(messages.SERVER_STARTING)

		err := common.Ternaryf(
			useSSL,
			func() error {
				return httpServer.ListenAndServeTLS(certFlag, keyFlag)
			},
			func() error {
				return httpServer.ListenAndServe()
			},
		)

		if !errors.Is(err, http.ErrServerClosed) {
			s.Logger().
				WithError(err).
				Error(messages.SERVER_ERROR)

			// Shut ourself down
			changes <- svc.Status{State: svc.StopPending}
			cancel()
		}
	}()

	return httpServer
}

// apiKeyMiddleware is a Gin middleware that checks for a valid API key
// in the "X-Api-Key" header of incoming requests.
// If the API key is missing or invalid, it aborts the request with a 403 Forbidden response.
func apiKeyMiddleware(logger *logrus.Logger, apiKey string) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		needApiKey := func() bool {
			path := ctx.Request.URL.Path
			for _, p := range []string{"/", "/swagger", "/healthz"} {
				if p == path || strings.HasPrefix(path, p) {
					return false
				}
			}
			return true
		}()

		if needApiKey {
			key := ctx.Request.Header.Get("X-Api-Key")

			if key == "" || !strings.EqualFold(key, apiKey) {
				remoteAddr := func() string {
					switch {
					case ctx.ClientIP() != "":
						return ctx.ClientIP()
					default:
						return "<unknown>"
					}
				}()

				logger.
					WithField("endpoint", ctx.Request.URL.String()).
					WithField("source", remoteAddr).
					Warn("Access was denied")

				ctx.AbortWithStatusJSON(http.StatusForbidden, &rest.Error{
					Code:    codes.PermissionDenied,
					Message: "Invalid API key",
				})
				return
			}
		}

		ctx.Next()
	}
}
