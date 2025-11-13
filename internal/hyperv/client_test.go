package hyperv

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

// Test all code paths of the execute method called
// by all Hyper-V client methods

func (s *ClientTestSuite) TestExecute() {

	expected := &rest.GetVolumeResponse{
		ID:   "id",
		Size: 1,
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusCreated,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer(
					s.MustMarshalJSON(expected),
				),
			},
		},
		nil,
	)

	actual, err := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET")

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *ClientTestSuite) TestExecuteNoResult() {

	expected := &noResult{}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusCreated,
			Body: &closeableBuffer{
				buf: &bytes.Buffer{},
			},
		},
		nil,
	)

	actual, err := apiCall[*noResult](context.Background(), s.client, "test", s.mustRequestURL(), "GET")

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *ClientTestSuite) TestExecuteFailOnDo() {

	underlyingError := errors.New("an error")

	req := s.mustRequestURL()

	expected := &url.Error{
		Op:  "GET",
		URL: req.String(),
		Err: underlyingError,
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(nil, expected)

	_, actual := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", req, "GET")
	s.Require().Error(actual)
	s.Require().Contains(actual.Error(), "error making request")
	urlErr := &url.Error{}
	s.Require().ErrorAs(actual, &urlErr)
	s.Require().Equal(expected, urlErr)
}

func (s *ClientTestSuite) TestExecuteFailOnStatus() {

	expected := &rest.Error{
		Code:    codes.Internal,
		Message: "an error",
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer(
					s.MustMarshalJSON(expected),
				),
			},
		},
		nil,
	)

	_, actual := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET")

	s.Require().Error(actual)
	restErr := &rest.Error{}
	s.Require().ErrorAs(actual, &restErr)
	s.Require().Equal(expected, restErr)
}

func (s *ClientTestSuite) TestExecuteFailUnmarshallingResponse() {

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusCreated,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer([]byte("this cannot be unmarshaled")),
			},
		},
		nil,
	)

	_, err := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET")

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "error unmarshaling response data")
}

func (s *ClientTestSuite) TestExecuteFailUnmashallingError() {

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer([]byte("this cannot be unmarshaled")),
			},
		},
		nil,
	)

	_, err := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET")

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "error unmarshaling error response")
}

func (s *ClientTestSuite) TestExecuteFailOnTimeout() {

	s.mockHttp.EXPECT().Do(mock.Anything).RunAndReturn(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, err := apiCall[*rest.GetVolumeResponse](ctx, s.client, "test", s.mustRequestURL(), "GET")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "error making request: context deadline exceeded")
}

var errRead = errors.New("read error")

type readerWithError struct{}

func (readerWithError) Read([]byte) (int, error) {
	return 0, errRead
}

func (readerWithError) Close() error {
	return nil
}

func (s *ClientTestSuite) TestExecuteFailReadResponseBody() {

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       &readerWithError{},
		},
		nil,
	)

	_, err := apiCall[*rest.GetVolumeResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET")
	s.Require().Error(err)
	s.Require().ErrorIs(err, errRead)
	s.Require().Contains(err.Error(), "error reading result")
}
