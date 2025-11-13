package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestExpandVolume() {

	var (
		id   = uuid.NewString()
		size = int64(constants.MiB * 10)
	)

	expected := &rest.ExpandVolumeResponse{
		NodeExpansionRequired: true,
		CapacityBytes:         size,
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

	actual, err := s.client.ExpandVolume(context.Background(), id, size)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}
