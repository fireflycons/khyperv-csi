package common

// Ternary performs a simple ternary operation on condition.
// Note that both trueVal and falseVal will be evaluated.
func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// Ternaryf performs a ternary operation on condition.
// trueVal and falseVal are functions rather than values
// therefore it is guaranteed that only one of them will be called.
func Ternaryf[T any](condition bool, trueVal, falseVal func() T) T {
	if condition {
		return trueVal()
	}
	return falseVal()
}
