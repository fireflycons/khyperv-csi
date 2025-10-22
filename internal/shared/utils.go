package shared

import (
	"strconv"
	"strings"

	"github.com/fireflycons/hypervcsi/internal/constants"
)

func FormatBytes(inputBytes int64) string {
	output := float64(inputBytes)
	unit := ""

	switch {
	case inputBytes >= constants.TiB:
		output /= constants.TiB
		unit = "Ti"
	case inputBytes >= constants.GiB:
		output /= constants.GiB
		unit = "Gi"
	case inputBytes >= constants.MiB:
		output /= constants.MiB
		unit = "Mi"
	case inputBytes >= constants.KiB:
		output /= constants.KiB
		unit = "Ki"
	case inputBytes == 0:
		return "0"
	}

	result := strconv.FormatFloat(output, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}
