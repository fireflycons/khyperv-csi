//go:build windows

package powershell

import (
	"regexp"

	"google.golang.org/grpc/codes"
)

var errCodeRx = regexp.MustCompile(`^[A-Z_]+`)

func extractCsiErrorCode(stderr string) codes.Code {
	// Now look for a canonical gRPC error string at start of the message
	if matches := errCodeRx.FindStringSubmatch(stderr); len(matches) > 0 {
		var c codes.Code
		if err := c.UnmarshalJSON([]byte(`"` + matches[0] + `"`)); err == nil {
			return c
		}
	}

	return codes.Unknown
}
