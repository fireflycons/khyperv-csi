package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestListVolume() {

	expected := &rest.ListVolumesResponse{
		Volumes: []*models.GetVHDResponse{
			{
				DiskIdentifier: "1",
			},
			{
				DiskIdentifier: "2",
			},
		},
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

	actual, err := s.client.ListVolumes(context.Background(), 0, "")

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *ClientTestSuite) TestListVolumesNegativeEntriesIsError() {

	_, err := s.client.ListVolumes(context.Background(), -1, "")
	s.Require().Error(err)
	s.Require().ErrorIs(err, errNegativeValue)
}
