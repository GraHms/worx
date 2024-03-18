package router

type fieldType map[string]interface{}

func fieldSelector(fields []string, data fieldType) (fieldType, *Err) {
	if len(fields) == 0 {
		return data, nil
	}
	result := make(map[string]interface{})
	for _, field := range fields {

		if value, ok := data[field]; ok {
			result[field] = value
			continue
		}
		perr := Err{
			ErrCode:    "INVALID_FIELD_ERROR",
			ErrReason:  "The field <" + field + "> does not exist",
			Message:    "Invalid field <" + field + ">",
			StatusCode: 400,
		}
		return nil, &perr

	}
	return result, nil
}

func removeNilPointers(data map[string]interface{}) {
	for key, value := range data {
		if value == nil {
			delete(data, key)
			continue
		}

		switch v := value.(type) {
		case map[string]interface{}:
			removeNilPointers(v)
			if len(v) == 0 {
				delete(data, key)
			}
		case []interface{}:
			newValue := removeNilPointersFromArray(v)
			data[key] = newValue
			if len(newValue) == 0 {
				delete(data, key)
			}

		}
	}
}

func removeNilPointersFromArray(arr []interface{}) []interface{} {
	for i := 0; i < len(arr); i++ {
		if arr[i] == nil {
			arr = append(arr[:i], arr[i+1:]...)
			i--
		} else if m, ok := arr[i].(map[string]interface{}); ok {
			removeNilPointers(m)
			if len(m) == 0 {
				arr = append(arr[:i], arr[i+1:]...)
				i--
			}
		} else if subArr, ok := arr[i].([]interface{}); ok {
			subArr = removeNilPointersFromArray(subArr)
			arr[i] = subArr
			if len(subArr) == 0 {
				arr = append(arr[:i], arr[i+1:]...)
				i--
			}
		}
	}
	return arr
}
