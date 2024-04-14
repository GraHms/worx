package router

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Map map[string]any
type OpenAPI struct {
	swagger     Map
	title       string
	version     string
	description string
	paths       Map
	endpoints   map[string]*Endpoint
}

func NewOpenAPI(title, version, description string) *OpenAPI {
	swagger := make(Map)
	swagger["openapi"] = "3.0.0"
	swagger["info"] = Map{
		"title":       title,
		"version":     version,
		"description": description,
	}
	return &OpenAPI{
		swagger:     swagger,
		title:       title,
		version:     version,
		description: description,
		paths:       make(Map),
	}
}

func (o *OpenAPI) SetEndpoints(endpoints map[string]*Endpoint) *OpenAPI {
	o.endpoints = endpoints
	return o
}

func (o *OpenAPI) Build() (Map, error) {
	if len(o.endpoints) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	for _, endpoint := range o.endpoints {
		o.paths[endpoint.Path] = o.buildPathItem(endpoint)
	}
	o.swagger["paths"] = o.paths
	return o.swagger, nil
}

func (o *OpenAPI) buildPathItem(endpoint *Endpoint) Map {
	pathItem := make(Map)
	path := endpoint.Path

	for _, method := range endpoint.Methods {
		pathItem[strings.ToLower(method.HTTPMethod)] = o.buildOperation(method)
	}

	o.paths[path] = pathItem
	return pathItem
}

func (o *OpenAPI) buildOperation(method Method) Map {
	statusCode := "200"
	if method.StatusCode != nil {
		statusCode = strconv.Itoa(*method.StatusCode)
	}

	operation := Map{
		"responses": Map{
			statusCode: o.buildResponse(),
		},
	}

	if method.HTTPMethod != "GET" && method.Request != nil {
		operation["requestBody"] = o.buildRequestBody(method.Request)
	}

	if method.Response != nil {
		schema := Schema{}
		operation["responses"].(Map)["200"].(Map)["content"].(Map)["application/json"].(Map)["schema"] = schema.Build(method.Response, "response")
	}
	tags := method.Tags

	tags = append(method.Configs.Tags)
	if tags != nil {
		operation["tags"] = tags
	}

	operation["description"] = method.Description
	operation["summary"] = method.Configs.Name

	parameters := o.buildParameters(method.Configs.AllowedHeaders, method.Configs.AllowedParams, method.Configs.PathParams)
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}

	return operation
}

func (o *OpenAPI) buildErrResponse(code, reason, message string) Map {
	return Map{
		"description": message,
		"content": Map{
			"application/json": Map{
				"schema": Map{
					"type": "object",
					"example": Map{
						"code":    code,
						"reason":  reason,
						"message": message,
					},
				},
			},
		},
	}
}

func (o *OpenAPI) buildResponse() Map {
	return Map{
		"description": "Successful operation",
		"content": Map{
			"application/json": Map{
				"schema": Map{
					"type": "object",
				},
			},
		},
	}
}

func (o *OpenAPI) buildRequestBody(request interface{}) Map {
	s := Schema{}
	requestSchema := s.Build(request, "request")
	return Map{
		"required": true,
		"content": Map{
			"application/json": Map{
				"schema": requestSchema,
			},
		},
	}
}

func (o *OpenAPI) buildParameters(headers, queryParams, pathParams []AllowedFields) []Map {
	parameters := make([]Map, 0)

	for _, header := range headers {
		parameters = append(parameters, o.buildHeaderParameter(header))
	}

	for _, param := range queryParams {
		parameters = append(parameters, o.buildQueryParamParameter(param))
	}

	for _, param := range pathParams {
		parameters = append(parameters, o.buildPathParamParameter(param))
	}

	return parameters
}

func (o *OpenAPI) buildHeaderParameter(header AllowedFields) Map {
	return o.buildParam(header, "header")
}

func (o *OpenAPI) buildQueryParamParameter(param AllowedFields) Map {
	return o.buildParam(param, "query")
}

func (o *OpenAPI) buildPathParamParameter(param AllowedFields) Map {
	return o.buildParam(param, "path")
}

func (o *OpenAPI) buildParam(param AllowedFields, t string) Map {
	return Map{
		"in":          t,
		"name":        param.Name,
		"description": param.Description,
		"required":    param.Required,
		"schema": Map{
			"type": "string",
		},
	}
}

type Schema struct{}

func (sc *Schema) Build(input interface{}, structType string) map[string]interface{} {

	if reflect.TypeOf(input).Kind() == reflect.Ptr {
		input = reflect.ValueOf(input).Elem().Interface()
	}
	t := reflect.TypeOf(input)

	schema := make(map[string]interface{})
	schema["type"] = "object"

	properties := make(map[string]interface{})
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := sc.getFieldName(field)
		if len(fieldName) == 0 {
			continue
		}
		fieldName = field.Tag.Get("json")
		required = sc.checkRequired(field, required)
		if len(field.Tag.Get("binding")) > 0 {
			binding := field.Tag.Get("binding")
			switch binding {
			case "ignore":
				if structType == "request" {
					continue
				}

			}
		}

		fieldSchema := sc.buildFieldSchema(field, structType)

		properties[fieldName] = fieldSchema

		required = sc.checkRequired(field, required)
	}

	schema["properties"] = properties

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema

}

func (sc *Schema) buildFieldSchema(field reflect.StructField, sType string) map[string]interface{} {
	fieldSchema := make(map[string]interface{})

	fieldName := field.Tag.Get("json")
	if fieldName == "" {
		fieldName = field.Name
	}

	fieldSchema["name"] = fieldName
	description := field.Tag.Get("description")
	if description != "" {
		fieldSchema["description"] = description
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	switch fieldType.Kind() {

	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			fieldSchema["type"] = "string"
			fieldSchema["format"] = "date-time"

		} else {
			nestedSchema := sc.Build(reflect.New(fieldType).Elem().Interface(), sType)
			fieldSchema["type"] = "object"
			fieldSchema["properties"] = nestedSchema["properties"]
			// If nested struct has required fields, include them in the parent schema
			if requiredFields, ok := nestedSchema["required"]; ok {
				if required, ok := requiredFields.([]string); ok {
					fieldSchema["required"] = required
				}
			}
		}
	case reflect.Slice:

		sliceType := fieldType.Elem()
		if sliceType.Kind() == reflect.Ptr {

			sliceType = sliceType.Elem()
		}
		if sliceType.Kind() == reflect.Struct {

			nestedSchema := sc.Build(reflect.New(sliceType).Elem().Interface(), sType)
			fieldSchema["type"] = "array"
			fieldSchema["items"] = nestedSchema
		} else if sliceType.Kind() == reflect.Slice {

			nestedSchema := sc.buildNestedListSchema(sliceType)
			fieldSchema["type"] = "array"
			fieldSchema["items"] = nestedSchema
		} else {

			fieldSchema["type"] = "array"
			fieldSchema["items"] = sc.getPrimitiveTypeSchema(sliceType)
		}
	default:
		fieldSchema = sc.getPrimitiveTypeSchema(fieldType)
		enums := field.Tag.Get("enums")
		if enums != "" {
			enumValues := strings.Split(enums, ",")
			fieldSchema["enum"] = enumValues
		}
	}

	regex := field.Tag.Get("regex")
	if regex != "" {

		fieldSchema["pattern"] = regex
	}

	example := field.Tag.Get("example")
	if example != "" {

		fieldSchema["example"] = example
	}

	return fieldSchema
}

func (sc *Schema) getPrimitiveTypeSchema(fieldType reflect.Type) map[string]interface{} {
	schema := make(map[string]interface{})
	switch fieldType.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema["type"] = "integer"
		schema["format"] = "int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
		schema["format"] = "uint64"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
		schema["format"] = "float"
	case reflect.Bool:
		schema["type"] = "boolean"
	default:
		schema["type"] = "string"

	}
	return schema
}

func (sc *Schema) buildNestedListSchema(sliceType reflect.Type) map[string]interface{} {
	nestedSchema := make(map[string]interface{})
	nestedSchema["type"] = "array"
	nestedSchema["items"] = sc.getPrimitiveTypeSchema(sliceType.Elem())
	return nestedSchema
}
func (sc *Schema) getFieldName(field reflect.StructField) string {
	if len(field.Tag.Get("json")) > 0 {
		return field.Tag.Get("json")
	}
	return field.Name
}
func (sc *Schema) checkRequired(field reflect.StructField, required []string) []string {
	if binding := field.Tag.Get("binding"); len(binding) > 0 && binding == "required" {
		required = append(required, sc.getFieldName(field))
	}
	return required
}

const SwagTempl = `
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
