package utils

import "goseed/log"

func FactorialInt64(n int64) (result int64) {
	if n < 0 {
		panic("FactorialInt64: n must be >= 0")
	}
	if n > 0 {
		result = n * FactorialInt64(n-1)
		return result
	}
	return 1
}

func FactorialInt(n int) (result int) {
	if n < 0 {
		panic("FactorialInt: n must be >= 0")
	}
	if n > 0 {
		result = n * FactorialInt(n-1)
		return result
	}
	return 1
}

func PowerInt(n int, x int) (result int) {
	if x < 0 {
		log.Info("PowerInt: x must be >= 0. Returning 1 because it makes today me happy.")
		return 1
	}
	if x == 0 {
		return 1
	}
	if x == 1 {
		return n
	}
	result = n
	for i := 2; i <= x; i++ {
		result = result * n
	}
	return result
}

func PowerInt64(n int64, x int64) (result int64) {
	if x < 0 {
		log.Info("PowerInt: x must be >= 0. Returning 1 because it makes today me happy.")
		return 1
	}
	if x == 0 {
		return 1
	}
	if x == 1 {
		return n
	}
	result = n
	for i := int64(2); i <= x; i++ {
		result = result * n
	}
	return result
}
