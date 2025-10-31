package common

import (
	"encoding/json"

	"github.com/stretchr/testify/suite"
)

type SuiteBase struct {
	suite.Suite
}

func (s *SuiteBase) MustMarshalJSON(data any) []byte {
	b, err := json.Marshal(data)
	s.Require().NoError(err, "cannot marshal JSON")
	return b
}
