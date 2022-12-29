package main

import (
	"reflect"
)

func sliceContains[T any](arr []T, thing T) bool {
	for _, v := range arr {
		if reflect.DeepEqual(v, thing) {
			return true
		}
	}

	return false
}
