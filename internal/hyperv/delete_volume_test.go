package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestDeleteVolume() {

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body: &closeableBuffer{
				buf: &bytes.Buffer{},
			},
		},
		nil,
	)

	err := s.client.DeleteVolume(context.Background(), "0000")
	s.Require().NoError(err)
}
