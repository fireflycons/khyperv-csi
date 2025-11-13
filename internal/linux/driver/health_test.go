//go:build linux

package driver

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/fireflycons/hypervcsi/internal/hyperv"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"google.golang.org/grpc/codes"
)

func (s *driverTestSuite) TestHVHealthChecker_Name() {
	c := hvHealthChecker{}
	s.Require().Equal(hvHealthCheckerName, c.Name())
}

func (s *driverTestSuite) TestHVHealthCheker_Check() {
	c := hvHealthChecker{}

	s.Run("healthy backend", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(s.MustMarshalJSON(rest.HealthyResponse{
				Status: "ok",
			}))
			s.Require().NoError(err)
		}))
		defer ts.Close()

		client, err := hyperv.NewClient(ts.URL, &http.Client{}, "a-key", nil)
		s.Require().NoError(err)

		c.client = client
		s.Require().NoError(c.Check(context.Background()))
	})

	s.Run("unhealthy backend", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write(s.MustMarshalJSON(rest.Error{
				Code:    codes.Internal,
				Message: "unhealthy",
			}))
			s.Require().NoError(err)
		}))
		defer ts.Close()

		client, err := hyperv.NewClient(ts.URL, &http.Client{}, "a-key", nil)
		s.Require().NoError(err)

		c.client = client

		err = c.Check(context.Background())
		s.Require().Error(err)
		restErr := &rest.Error{}
		s.Require().ErrorAs(err, &restErr)
		s.Require().Equal(codes.Internal, restErr.Code)
	})
}
