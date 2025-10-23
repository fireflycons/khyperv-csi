package hyperv

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	mockHttp *mockhttpClient
	client   client
}

var (
	_ suite.BeforeTest = (*ClientTestSuite)(nil)
)

func (s *ClientTestSuite) BeforeTest(_, _ string) {
	s.mockHttp = newMockhttpClient(s.T())
	s.client = client{
		httpClient: s.mockHttp,
		addr:       s.mustParseUrl("http://localhost/"),
	}
}

func TestClientPackage(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type closeableBuffer struct {
	buf *bytes.Buffer
}

func (c *closeableBuffer) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}

func (closeableBuffer) Close() error {
	return nil
}

var _ io.ReadCloser = (*closeableBuffer)(nil)

func (s *ClientTestSuite) mustParseUrl(addr string) *url.URL {
	u, err := url.Parse(addr)
	s.Require().NoError(err, "cannot parse URL")
	return u
}

func (s *ClientTestSuite) mustMarshalJSON(data any) []byte {
	b, err := json.Marshal(data)
	s.Require().NoError(err, "cannot marshal JSON")
	return b
}

func (s *ClientTestSuite) mustRequestURL() *url.URL {
	u, err := url.Parse("http://localhost/")
	s.Require().NoError(err, "mustRequestURL")
	return u
}
