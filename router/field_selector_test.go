package router

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type ObjWithNil struct {
	Foo   string      `json:"foo"`
	Baz   *string     `json:"baz"`
	Grhms []*string   `json:"grhms"`
	More  *ObjWithNil `json:"more"`
}

func TestRemoveNilPointers(t *testing.T) {
	data := "foo"
	instance := &ObjWithNil{
		Foo:   "bar",
		Baz:   nil,
		Grhms: []*string{&data, nil},
		More: &ObjWithNil{
			Grhms: []*string{nil, nil},
		},
	}
	instanceByte, _ := json.Marshal(instance)
	input := map[string]interface{}{}
	_ = json.Unmarshal(instanceByte, &input)
	removeNilPointers(input)
	assert.NotNil(t, input["foo"])
}

func TestShouldRemoveNilPointers(t *testing.T) {
	// Test case 1: Nested map with nil value
	data := map[string]interface{}{
		"key1": nil,
		"key2": "value2",
		"key3": map[string]interface{}{
			"key4": nil,
			"key5": "value5",
		},
		"key6": []interface{}{
			"value6",
			nil,
			map[string]interface{}{
				"key7": nil,
				"key8": "value8",
			},
		},
	}

	removeNilPointers(data)

	assert.NotNil(t, data)

	// Test case 2: Empty map
	data = map[string]interface{}{}

	removeNilPointers(data)

	assert.NotNil(t, data)
	// Test case 3: Nested array with nil value
	data = map[string]interface{}{
		"key1": []interface{}{
			nil,
			"value2",
			map[string]interface{}{
				"key3": nil,
				"key4": "value4",
			},
		},
	}

	removeNilPointers(data)

	assert.NotNil(t, data)
}

func TestRemoveNilPointers_NestedMap(t *testing.T) {
	data := map[string]interface{}{
		"key1": nil,
		"key2": map[string]interface{}{
			"nestedKey1": nil,
			"nestedKey2": "value2",
		},
	}

	removeNilPointers(data)

	expected := map[string]interface{}{
		"key2": map[string]interface{}{
			"nestedKey2": "value2",
		},
	}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("Expected %v, got %v", expected, data)
	}
}
func TestRemoveNilPointers_EmptyNestedMap(t *testing.T) {
	data := map[string]interface{}{
		"key1": nil,
		"key2": map[string]interface{}{
			"nestedKey1": nil,
		},
	}

	removeNilPointers(data)

	expected := map[string]interface{}{}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("Expected %v, got %v", expected, data)
	}
}

func TestRemoveNilPointers_Slice(t *testing.T) {
	data := map[string]interface{}{
		"key1": nil,
		"key2": []interface{}{nil, "value1", nil},
	}

	removeNilPointers(data)

	expected := map[string]interface{}{
		"key2": []interface{}{"value1"},
	}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("Expected %v, got %v", expected, data)
	}
}

func TestRemoveNilPointers_EmptySlice(t *testing.T) {
	data := map[string]interface{}{
		"key1": nil,
		"key2": []interface{}{nil, nil},
	}

	removeNilPointers(data)

	expected := map[string]interface{}{}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("Expected %v, got %v", expected, data)
	}
}

func TestRemoveNilPointers_SliceWithMap(t *testing.T) {
	data := []interface{}{
		nil,
		map[string]interface{}{
			"nestedKey1": nil,
			"nestedKey2": "value2",
		},
	}

	result := removeNilPointersFromArray(data)

	expected := []interface{}{
		map[string]interface{}{
			"nestedKey2": "value2",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestRemoveNilPointers_EmptyMapInSlice(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{
			"nestedKey1": nil,
		},
	}

	result := removeNilPointersFromArray(data)

	expected := []interface{}{}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestRemoveNilPointers_SliceWithinSlice(t *testing.T) {
	data := []interface{}{
		nil,
		[]interface{}{nil, "value1", nil},
	}

	result := removeNilPointersFromArray(data)

	expected := []interface{}{
		[]interface{}{"value1"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestRemoveNilPointers_EmptySliceInSlice(t *testing.T) {
	data := []interface{}{
		[]interface{}{nil, nil},
	}

	result := removeNilPointersFromArray(data)

	expected := []interface{}{}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestRemoveNilPointer(t *testing.T) {
	t.Run("Should retain non-nil fields in an API response", func(t *testing.T) {
		data := "active"
		instance := &ObjWithNil{
			Foo:   "John",
			Baz:   nil,
			Grhms: []*string{&data, nil},
			More: &ObjWithNil{
				Grhms: []*string{nil, nil},
			},
		}
		instanceByte, _ := json.Marshal(instance)
		input := map[string]interface{}{}
		_ = json.Unmarshal(instanceByte, &input)
		removeNilPointers(input)
		assert.NotNil(t, input["foo"])
	})

	t.Run("Should handle API response with nested user profiles", func(t *testing.T) {
		data := map[string]interface{}{
			"current_user": nil,
			"friend": map[string]interface{}{
				"firstName": nil,
				"lastName":  "Doe",
			},
			"activities": []interface{}{
				"Running",
				nil,
				map[string]interface{}{
					"task":   nil,
					"status": "completed",
				},
			},
		}

		removeNilPointers(data)

		expected := map[string]interface{}{
			"friend": map[string]interface{}{
				"lastName": "Doe",
			},
			"activities": []interface{}{
				"Running",
				map[string]interface{}{
					"status": "completed",
				},
			},
		}
		assert.Equal(t, expected, data)
	})

	t.Run("Should clean API response with an empty user list", func(t *testing.T) {
		data := map[string]interface{}{
			"users": []interface{}{
				nil,
				"Anna",
				map[string]interface{}{
					"firstName": nil,
					"lastName":  "Smith",
				},
			},
		}

		removeNilPointers(data)

		expected := map[string]interface{}{
			"users": []interface{}{
				"Anna",
				map[string]interface{}{
					"lastName": "Smith",
				},
			},
		}
		assert.Equal(t, expected, data)
	})

	t.Run("Should handle API response with nested activities", func(t *testing.T) {
		data := map[string]interface{}{
			"activities": nil,
			"userActivities": []interface{}{
				nil,
				"Swimming",
				nil,
			},
		}

		removeNilPointers(data)

		expected := map[string]interface{}{
			"userActivities": []interface{}{"Swimming"},
		}

		assert.Equal(t, expected, data)
	})

	t.Run("Should handle a user's activities list containing nested tasks", func(t *testing.T) {
		data := []interface{}{
			nil,
			map[string]interface{}{
				"activityName": nil,
				"status":       "ongoing",
			},
		}

		result := removeNilPointersFromArray(data)

		expected := []interface{}{
			map[string]interface{}{
				"status": "ongoing",
			},
		}

		assert.Equal(t, expected, result)
	})

	t.Run("Should handle nested user lists within main list", func(t *testing.T) {
		data := []interface{}{
			nil,
			[]interface{}{nil, "Sarah", nil},
		}

		result := removeNilPointersFromArray(data)

		expected := []interface{}{
			[]interface{}{"Sarah"},
		}

		assert.Equal(t, expected, result)
	})

}

func TestFieldSelector(t *testing.T) {
	data := fieldType{
		"name": "John",
		"address": fieldType{
			"street": "123 Main St",
			"city":   "Metropolis",
		},
		"friends": []interface{}{
			fieldType{
				"name": "Jane",
			},
			fieldType{
				"name": "Doe",
			},
		},
	}

	t.Run("should select top-level fields", func(t *testing.T) {
		fields := []string{"name"}
		result, err := fieldSelector(fields, data)
		if err != nil {
			t.Fatal(err)
		}
		if result["name"] != "John" {
			t.Errorf("Expected 'John', got '%v'", result["name"])
		}
	})

	t.Run("should return error for non-existent fields", func(t *testing.T) {
		fields := []string{"address.zip"}
		_, err := fieldSelector(fields, data)
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
	})
}
