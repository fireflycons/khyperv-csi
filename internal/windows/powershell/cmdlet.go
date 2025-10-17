//go:build windows

package powershell

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fireflycons/hypervcsi/internal/constants"
)

type ErrUnsupportedDataType struct {
	Message string
}

func (e *ErrUnsupportedDataType) Error() string {
	return e.Message
}

func (*ErrUnsupportedDataType) Is(err error) bool {
	_, ok := err.(*ErrUnsupportedDataType)
	return ok
}

type Cmdlet struct {
	Name        string
	FullCommand string
	Args        map[string]any
	Err         error
}

func NewCmdlet(cmdlet string, args map[string]any) Cmdlet {
	c := Cmdlet{
		Name: cmdlet,
		Args: args,
	}

	f, err := buildCmdlet(cmdlet, args)
	c.FullCommand = f
	c.Err = err
	return c
}

func buildCmdlet(cmdlet string, args map[string]any) (string, error) {
	sb := strings.Builder{}
	sb.Grow(constants.KiB)

	sb.WriteString(cmdlet)

	for arg, value := range args {

		sb.WriteRune(' ')
		if !strings.HasPrefix(arg, "-") {
			sb.WriteRune('-')
		}
		sb.WriteString(arg)

		if value == nil {
			continue
		}

		sb.WriteRune(' ')
		if s, err := formatValue(value); err == nil {
			sb.WriteString(s)
		} else {
			return "", err
		}
	}

	return sb.String(), nil
}

func formatValue(value any) (string, error) { //nolint:gocyclo // needs a large numer of cases

	switch v := value.(type) {
	case string:
		return formatString(v), nil
	case []string:
		sa := make([]string, len(v))
		for i := range v {
			sa[i] = formatString(v[i])
		}
		return strings.Join(sa, ","), nil
	case bool:
		return func() string {
			if v {
				return "$true"
			}
			return "$false"
		}(), nil

	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case []int:
		return formatList(v), nil
	case []int8:
		return formatList(v), nil
	case []int16:
		return formatList(v), nil
	case []int32:
		return formatList(v), nil
	case []int64:
		return formatList(v), nil
	case []uint:
		return formatList(v), nil
	case []uint8:
		return formatList(v), nil
	case []uint16:
		return formatList(v), nil
	case []uint32:
		return formatList(v), nil
	case []uint64:
		return formatList(v), nil
	case []float32:
		return formatList(v), nil
	case []float64:
		return formatList(v), nil
	}

	return "", &ErrUnsupportedDataType{
		Message: fmt.Sprintf("unsupported datatype %T", value),
	}
}

func formatString(s string) string {

	dQuote := strings.Contains(s, "\"")
	ws := strings.Contains(s, " ")
	switch {
	case ws || dQuote:
		return "'" + s + "'"
	default:
		return "\"" + s + "\""
	}
}

func formatList[T any](values []T) string {
	arr := make([]string, len(values))

	for i, v := range values {
		arr[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(arr, ",")
}
