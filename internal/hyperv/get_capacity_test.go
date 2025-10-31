package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestGetCapacity() {

	expected := &rest.GetCapacityResponse{
		AvailableCapacity: 2 * constants.TiB,
		MinimumVolumeSize: constants.MinimumVolumeSizeInBytes,
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

	actual, err := s.client.GetCapacity(context.Background())

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}
