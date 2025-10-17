//go:build windows

package powershell

import (
	"errors"

	"google.golang.org/grpc/codes"
)

func (s *PowershellTestSuite) TestExtractCsiErrorCode() {
	tests := []struct {
		code string
		want codes.Code
	}{
		{
			code: "OK",
			want: codes.OK,
		},
		{
			//nolint:misspell // this is how it is spelled in the grpc codes package
			code: "CANCELLED",
			want: codes.Canceled,
		},
		{
			code: "UNKNOWN",
			want: codes.Unknown,
		},
		{
			code: "INVALID_ARGUMENT",
			want: codes.InvalidArgument,
		},
		{
			code: "DEADLINE_EXCEEDED",
			want: codes.DeadlineExceeded,
		},
		{
			code: "NOT_FOUND",
			want: codes.NotFound,
		},
		{
			code: "ALREADY_EXISTS",
			want: codes.AlreadyExists,
		},
		{
			code: "PERMISSION_DENIED",
			want: codes.PermissionDenied,
		},
		{
			code: "RESOURCE_EXHAUSTED",
			want: codes.ResourceExhausted,
		},
		{
			code: "FAILED_PRECONDITION",
			want: codes.FailedPrecondition,
		},
		{
			code: "ABORTED",
			want: codes.Aborted,
		},
		{
			code: "OUT_OF_RANGE",
			want: codes.OutOfRange,
		},
		{
			code: "UNIMPLEMENTED",
			want: codes.Unimplemented,
		},
		{
			code: "INTERNAL",
			want: codes.Internal,
		},
		{
			code: "UNAVAILABLE",
			want: codes.Unavailable,
		},
		{
			code: "DATA_LOSS",
			want: codes.DataLoss,
		},
		{
			code: "UNAUTHENTICATED",
			want: codes.Unauthenticated,
		},
		{
			code: "NOT_A_VALID_CODE",
			want: codes.Unknown,
		},
	}

	runner := s.runner

	for _, tt := range tests {
		s.Run(tt.code, func() {

			err := runner.Run(
				NewCmdlet(
					"Test-PVException",
					map[string]any{
						"CanonicalCsiError": tt.code,
					},
				),
			)

			s.Require().Error(err)

			targetError := &RunnerError{}
			s.Require().True(errors.As(err, &targetError))
			s.Require().Equal(tt.want, targetError.Code)
		})
	}
}
