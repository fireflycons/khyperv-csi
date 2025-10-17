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

func (s *ClientTestSuite) TestCreateVolume() {

	var (
		id   = uuid.NewString()
		size = int64(constants.MiB * 10)
	)

	expected := &rest.CreateVolumeResponse{
		ID:   id,
		Size: size,
	}

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusCreated,
			Body: &closeableBuffer{
				buf: bytes.NewBuffer(
					s.mustMarshalJSON(expected),
				),
			},
		},
		nil,
	)

	actual, err := s.client.CreateVolume(context.Background(), "test", size)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *ClientTestSuite) TestCreateVolumeNegativeSizeIsError() {

	_, err := s.client.CreateVolume(context.Background(), "test", -1)
	s.Require().Error(err)
	s.Require().ErrorIs(err, errNegativeValue)
}
