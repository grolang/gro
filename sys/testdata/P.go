package p

func f() {
	var x interface{}
	switch x.(type) {
	}
	switch x.(type) {
	}

	switch x.(type) {
	case int:
	}
	switch x.(type) {
	case int:
	}

	switch x.(type) {
	case []int:
	}

	switch t := x.(type) {
	default:
		_ = t
	}

}
