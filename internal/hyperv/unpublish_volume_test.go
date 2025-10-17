package hyperv

import (
	"bytes"
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *ClientTestSuite) TestUnpublishVolume() {

	var (
		nodeId = uuid.NewString()
		volId  = uuid.NewString()
	)

	s.mockHttp.EXPECT().Do(mock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body: &closeableBuffer{
				buf: &bytes.Buffer{},
			},
		},
		nil,
	)

	err := s.client.UnpublishVolume(context.Background(), volId, nodeId)
	s.Require().NoError(err)
}
