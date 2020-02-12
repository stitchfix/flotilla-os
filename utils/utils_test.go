package utils

import (
	"reflect"
	"testing"
)

func TestMergeMaps_Simple(t *testing.T) {
	mapA := map[string]interface{}{
		"A": "aaa",
		"B": "bbb",
		"C": "ccc",
	}
	mapB := map[string]interface{}{
		"B": "xxx",
		"D": "ddd",
	}

	expectedMapA := map[string]interface{}{
		"A": "aaa",
		"B": "bbb",
		"C": "ccc",
		"D": "ddd",
	}

	err := MergeMaps(&mapA, mapB)

	if err != nil {
		t.Error("unable to merge maps")
	}

	if reflect.DeepEqual(mapA, expectedMapA) == false {
		t.Error("map merge unsuccessful")
	}
}

func TestMergeMaps_Nested(t *testing.T) {
	nestedAValue := "aaa"
	nestedCValue := "ccc"
	overrideNestedBVal := "zzzzzz"
	nestedD1Value := "d1"
	overrideNestedD1Value := "override_d1"
	overrideNestedD2Value := "override_d2"

	mapA := map[string]interface{}{
		"Nested": map[string]interface{}{
			"A": nestedAValue,
			"C": nestedCValue,
			"D": map[string]interface{}{
				"D1": nestedD1Value,
			},
		},
	}

	mapB := map[string]interface{}{
		"Nested": map[string]interface{}{
			"B": overrideNestedBVal,
			"D": map[string]interface{}{
				"D1": overrideNestedD1Value,
				"D2": overrideNestedD2Value,
			},
		},
	}

	// After merging, mapA should have its `B` value set. Additionally, mapA[D]
	// should have its D2 value set BUT its D1 value should not be overriden.
	expectedMapA := map[string]interface{}{
		"Nested": map[string]interface{}{
			"A": nestedAValue,
			"B": overrideNestedBVal,
			"C": nestedCValue,
			"D": map[string]interface{}{
				"D1": nestedD1Value,
				"D2": overrideNestedD2Value,
			},
		},
	}

	err := MergeMaps(&mapA, mapB)

	if err != nil {
		t.Error("unable to merge maps")
	}

	if reflect.DeepEqual(mapA, expectedMapA) == false {
		t.Error("map merge unsuccessful")
	}
}
