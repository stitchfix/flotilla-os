package utils

import (
	"reflect"

	"github.com/pkg/errors"
)

// StringSliceContains checks is a string slice contains a particular string.
func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// MergeMaps takes a pointer to a map (first arg) and map containing default
// values (second arg) and recursively sets values that exist in `b` but are
// not set in `a`. For existing values, it does not override those of `a` with
// those of `b`.
func MergeMaps(a *map[string]interface{}, b map[string]interface{}) error {
	return mergeMapsRecursive(a, b)
}

func mergeMapsRecursive(a *map[string]interface{}, b map[string]interface{}) error {
	for k, v := range b {
		// If the value is a map, check recursively.
		if reflect.TypeOf(v).Kind() == reflect.Map {
			if _, ok := (*a)[k]; !ok {
				(*a)[k] = v
			} else {
				aVal, ok := (*a)[k].(map[string]interface{})
				bVal, ok := v.(map[string]interface{})

				if !ok {
					return errors.New("unable to cast interface{} to map[string]interface{}")
				}

				if err := mergeMapsRecursive(&aVal, bVal); err != nil {
					return err
				}
			}
		} else {
			if _, ok := (*a)[k]; !ok {
				(*a)[k] = v
			}
		}
	}

	return nil
}
