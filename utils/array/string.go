package array

func Contain[T comparable](a []T, b T) bool {
	for _, v := range a {
		if v == b {
			return true
		}
	}
	return false
}

func Remove[T comparable](a []T, remove []T) []T {
	var results []T
	for _, v := range a {
		if !Contain[T](remove, v) {
			results = append(results, v)
		}
	}
	return results
}
