package worx

import (
	"fmt"
	"github.com/grahms/worx/router"
	"reflect"
	"strings"
)

func buildSwaggerJSON(name string) map[string]interface{} {
	swagger := make(map[string]interface{})
	swagger["openapi"] = "3.0.0"
	swagger["info"] = map[string]interface{}{
		"title":       name,
		"description": "Description of your API",
		"version":     "1.0.0",
	}

	paths := make(map[string]interface{})
	for _, endpoint := range router.Endpoints {
		pathItem := make(map[string]interface{})
		for _, method := range endpoint.Methods {
			operation := make(map[string]interface{})
			operation["responses"] = map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful operation",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
			}

			// Add request schema only if the method is not "GET"
			if method.HTTPMethod != "GET" && method.Request != nil {
				requestSchema := buildSchemaFromStruct(method.Request, "request")
				operation["requestBody"] = map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": requestSchema,
						},
					},
				}
			}

			if method.Response != nil {
				responseSchema := buildSchemaFromStruct(method.Response, "response")
				operation["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"] = responseSchema
			}

			// Add description if available
			if method.Description != "" {
				operation["description"] = method.Description
			}

			pathItem[strings.ToLower(method.HTTPMethod)] = operation
		}
		paths[endpoint.Path] = pathItem
	}

	swagger["paths"] = paths
	return swagger
}

func getValueOf(val interface{}) reflect.Value {
	return reflect.ValueOf(val)
}
func isPtr(value reflect.Value) bool {
	return value.Kind() == reflect.Ptr
}

func isList(value reflect.Type) bool {
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array
}

func getStructTypeFromList(listType reflect.Type) reflect.Type {

	if listType.Kind() == reflect.Slice || listType.Kind() == reflect.Array {
		// Get the element type of the slice/array
		elementType := listType.Elem()
		// If the element type is a struct, return it
		if elementType.Kind() == reflect.Struct {
			return elementType
		}
	}
	return nil
}

func buildSchemaFromStructByType(structType reflect.Type, schemaType string) map[string]interface{} {
	schema := make(map[string]interface{})

	switch schemaType {
	case "object":
		schema["type"] = "object"
		properties := make(map[string]interface{})
		required := make([]string, 0)

		// Iterate over the fields of the struct
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldName := field.Name

			// Check if the field is required based on struct tags
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				fieldName = jsonTag
			}
			if bindingTag := field.Tag.Get("binding"); bindingTag == "required" {
				required = append(required, fieldName)
			}

			// Determine the schema for the field
			fieldSchema := buildFieldFromSchema(field)

			// Add the field schema to the properties map
			properties[fieldName] = fieldSchema
		}

		// Add properties and required fields to the schema
		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}

	case "array":
		schema["type"] = "array"
		schema["items"] = buildSchemaFromStructByType(structType, "object")

	default:
		// Unsupported schema type
		fmt.Println("Unsupported schema type:", schemaType)
	}

	return schema
}

func buildSchemaFromStruct(s interface{}, structType string) map[string]interface{} {
	// Dereference pointer if s is a pointer'
	if reflect.TypeOf(s).Kind() == reflect.Ptr {
		s = reflect.ValueOf(s).Elem().Interface()
	}
	t := reflect.TypeOf(s)
	if structType := getStructTypeFromList(t); structType != nil {
		fmt.Printf("%v", structType)
		// If s is a slice or an array of structs, use the struct type to build the schema
		return buildSchemaFromStructByType(structType, "array")
	}

	schema := make(map[string]interface{})
	schema["type"] = "object"

	properties := make(map[string]interface{})
	required := make([]string, 0) // Slice to store required field names

	// Helper function to recursively build schema for nested structs
	buildNestedSchema := func(fieldValue reflect.Value) map[string]interface{} {

		if isPtr(fieldValue) {
			// If it's a pointer, get the value it points to
			fieldValue = fieldValue.Elem()
		}
		ftype := fieldValue.Type()
		if isList(ftype) {

			// If it's a list, build schema for its elements
			listSchema := map[string]interface{}{
				"type":  "array",
				"items": buildSchemaFromStruct(fieldValue.Type(), structType),
			}
			return listSchema
		} else if fieldValue.Kind() == reflect.Struct {
			// If it's a struct, recursively build schema
			return buildSchemaFromStruct(fieldValue.Interface(), structType)
		}
		// Otherwise, return an empty schema
		return map[string]interface{}{}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		field.Tag.Get("binding")
		// Check if the field is required
		if len(field.Tag.Get("json")) > 0 {
			fieldName = field.Tag.Get("json")
		}
		if len(field.Tag.Get("binding")) > 0 {
			binding := field.Tag.Get("binding")
			switch binding {
			case "required":
				required = append(required, fieldName)
			case "ignore":
				if structType == "request" {
					continue
				}

			}
		}

		// Check if the field is a list or a nested struct
		fieldValue := reflect.ValueOf(s).Field(i)
		if isList(fieldValue.Type()) || fieldValue.Kind() == reflect.Struct {
			properties[fieldName] = buildNestedSchema(fieldValue)
		} else {
			// Otherwise, build schema for the field as before
			properties[fieldName] = buildFieldFromSchema(field)
		}
	}

	schema["properties"] = properties

	// Add required fields to the schema if any
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func buildFieldFromSchema(field reflect.StructField) map[string]interface{} {
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	var jsonType string
	switch fieldType.Kind() {
	case reflect.String:
		jsonType = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		jsonType = "integer"
	case reflect.Float32, reflect.Float64:
		jsonType = "number"
	case reflect.Struct:
		// If the field is a struct, recursively build schema for it
		return buildSchemaFromStruct(reflect.New(fieldType), "request")
	case reflect.Slice, reflect.Array:
		// Get the element type of the slice/array
		elementType := fieldType.Elem()
		// If the element type is a struct, return it
		if elementType.Kind() == reflect.Struct {
			return buildSchemaFromStruct(reflect.New(fieldType).Elem().Interface(), "request")
		}

	default:
		jsonType = "object"
	}

	schema := map[string]interface{}{
		"type": jsonType,
	}

	// Check if the field has enums
	if enums := field.Tag.Get("enums"); len(enums) > 0 {
		enumValues := strings.Split(strings.TrimSpace(enums), ",")
		schema["enum"] = enumValues
	}

	return schema
}

var swagTempl = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.44.0/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.44.0/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      const spec = JSON.parse('{{.SwaggerJSON}}');
      const ui = SwaggerUIBundle({
        spec: spec,
        dom_id: '#swagger-ui',
      })
    }
  </script>
</body>
</html>
`
