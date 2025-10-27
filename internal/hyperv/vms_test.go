package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestListVms() {

	expected := &rest.ListVMResponse{
		VMs: []*rest.GetVMResponse{
			{
				Name:       "vm1",
				ID:         constants.ZeroUUID,
				Path:       "C:\\vms\\vm1",
				Generation: 2,
			},
		},
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer(
					s.MustMarshalJSON(expected),
				),
			},
		},
		nil,
	)

	actual, err := apiCall[*rest.ListVMResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET", "key")

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *ClientTestSuite) TestGetVms() {

	expected := &rest.GetVMResponse{

		Name:       "vm1",
		ID:         constants.ZeroUUID,
		Path:       "C:\\vms\\vm1",
		Generation: 2,
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer(
					s.MustMarshalJSON(expected),
				),
			},
		},
		nil,
	)

	actual, err := apiCall[*rest.GetVMResponse](context.Background(), s.client, "test", s.mustRequestURL(), "GET", "key")

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}
