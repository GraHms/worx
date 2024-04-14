package worx

import (
	"errors"
	"fmt"
	"github.com/grahms/worx/router"
	"reflect"
	"regexp"
	"strings"
)

type Map map[string]interface{}
type OpenAPI struct {
	swagger     Map
	title       string
	version     string
	description string
	paths       Map
	endpoints   map[string]*router.Endpoint
}

func New(title, version, description string) *OpenAPI {
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

func (o *OpenAPI) SetEndpoints(endpoints map[string]*router.Endpoint) *OpenAPI {
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

func (o *OpenAPI) buildPathItem(endpoint *router.Endpoint) Map {
	pathItem := make(Map)
	path := endpoint.Path
	re := regexp.MustCompile(`/:(\w+)(/|$)`)
	path = re.ReplaceAllString(path, "/{$1}/")
	re = regexp.MustCompile(`{(\w+)}/`)
	matches := re.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		paramName := match[1]
		if pathItem["parameters"] == nil {
			pathItem["parameters"] = []Map{}
		}
		pathItemParams := make(Map)
		pathItemParams["name"] = paramName
		pathItemParams["in"] = "path"
		pathItemParams["required"] = true
		pathItemParams["schema"] = Map{"type": "string"}
		pathItemParams["description"] = fmt.Sprintf("Path parameter %s", paramName)
		pathItem["parameters"] = append(pathItem["parameters"].([]Map), pathItemParams)
	}

	for _, method := range endpoint.Methods {
		pathItem[strings.ToLower(method.HTTPMethod)] = o.buildOperation(method)
	}

	o.paths[path] = pathItem
	return pathItem
}

func (o *OpenAPI) buildOperation(method router.Method) Map {
	operation := Map{
		"responses": Map{
			"200": o.buildResponse(),
		},
	}

	if method.HTTPMethod != "GET" && method.Request != nil {
		operation["requestBody"] = o.buildRequestBody(method.Request)
	}

	if method.Response != nil {
		schema := Schema{}
		operation["responses"].(Map)["200"].(Map)["content"].(Map)["application/json"].(Map)["schema"] = schema.Build(method.Response, "response")
	}
	operation["tags"] = method.Configs.Tags
	operation["description"] = method.Description
	operation["summary"] = method.Configs.Name

	parameters := o.buildParameters(method.Configs.AllowedHeaders, method.Configs.AllowedParams)
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}

	return operation
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

func (o *OpenAPI) buildParameters(headers []router.AllowedFields, queryParams []router.AllowedFields) []Map {
	parameters := make([]Map, 0)

	for _, header := range headers {
		parameters = append(parameters, o.buildHeaderParameter(header))
	}

	for _, param := range queryParams {
		parameters = append(parameters, o.buildQueryParamParameter(param))
	}

	return parameters
}

func (o *OpenAPI) buildHeaderParameter(header router.AllowedFields) Map {
	return Map{
		"in":          "header",
		"name":        header.Name,
		"description": header.Description,
		"required":    header.Required,
		"schema": Map{
			"type": "string",
		},
	}
}

func (o *OpenAPI) buildQueryParamParameter(param router.AllowedFields) Map {
	return Map{
		"in":          "query",
		"name":        param.Name,
		"description": param.Description,
		"required":    param.Required,
		"schema": Map{
			"type": "string",
		},
	}
}

func isPtr(value reflect.Value) bool {
	return value.Kind() == reflect.Ptr
}

func isList(value reflect.Type) bool {
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array
}

func (sc *Schema) buildSchemaFromStructByType(structType reflect.Type, schemaType string) map[string]interface{} {
	schema := make(map[string]interface{})

	switch schemaType {
	case "object":
		schema["type"] = "object"
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldName := field.Name

			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				fieldName = jsonTag
			}
			if bindingTag := field.Tag.Get("binding"); bindingTag == "required" {
				required = append(required, fieldName)
			}

			fieldSchema := sc.buildFieldFromSchema(field, schemaType)
			properties[fieldName] = fieldSchema
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}

	case "array":
		schema["type"] = "array"
		schema["items"] = sc.buildSchemaFromStructByType(structType, "object")

	default:
		fmt.Println("Unsupported schema type:", schemaType)
	}

	return schema
}

type Schema struct{}

func (sc *Schema) Build(input interface{}, structType string) map[string]interface{} {

	if reflect.TypeOf(input).Kind() == reflect.Ptr {
		input = reflect.ValueOf(input).Elem().Interface()
	}
	t := reflect.TypeOf(input)
	if stype := sc.getStructTypeFromList(t); stype != nil {
		return sc.buildSchemaFromStructByType(stype, "array")
	}

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

		// Check if the field is a list or a nested struct
		fieldValue := reflect.ValueOf(input).Field(i)
		if isList(fieldValue.Type()) || fieldValue.Kind() == reflect.Struct {
			properties[fieldName] = sc.buildNestedSchema(fieldValue, structType)
		} else {
			properties[fieldName] = sc.buildFieldFromSchema(field, structType)
		}
	}

	schema["properties"] = properties

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema

}
func (sc *Schema) buildNestedSchema(fieldValue reflect.Value, structType string) map[string]interface{} {
	if isPtr(fieldValue) {
		sc.buildNestedSchema(fieldValue.Elem(), structType)
	}
	ftype := fieldValue.Type()
	if isList(ftype) {
		listSchema := map[string]interface{}{
			"type":  "array",
			"items": sc.Build(fieldValue.Type(), structType),
		}
		return listSchema
	} else if fieldValue.Kind() == reflect.Struct {
		return sc.Build(fieldValue.Interface(), structType)
	}

	return map[string]interface{}{}
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

func (sc *Schema) buildSchemaForString(field reflect.StructField) map[string]interface{} {
	schema := map[string]interface{}{
		"type": "string",
	}
	if enums := field.Tag.Get("enums"); len(enums) > 0 {
		enumValues := strings.Split(strings.TrimSpace(enums), ",")
		schema["enum"] = enumValues
	}
	return schema
}

func (sc *Schema) buildSchemaForInteger() map[string]interface{} {
	return map[string]interface{}{
		"type": "integer",
	}
}

func (sc *Schema) buildSchemaForNumber() map[string]interface{} {
	return map[string]interface{}{
		"type": "number",
	}
}

func (sc *Schema) buildSchemaForStruct(structValue interface{}, structType string) map[string]interface{} {
	// Handle nested structs
	return sc.Build(structValue, structType)
}

func (sc *Schema) getStructTypeFromList(listType reflect.Type) reflect.Type {

	if listType.Kind() == reflect.Slice || listType.Kind() == reflect.Array {

		elementType := listType.Elem()

		if elementType.Kind() == reflect.Struct {
			return elementType
		}
	}
	return nil
}

//	func (sc *Schema) buildSchemaForArray(fieldType reflect.Type) map[string]interface{} {
//		return map[string]interface{}{
//			"type":  "array",
//			"items": sc.buildFieldFromSchema(reflect.New(fieldType).Elem().Interface()),
//		}
//	}
func (sc *Schema) buildFieldFromSchema(field reflect.StructField, inputType string) map[string]interface{} {
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
		return sc.Build(reflect.New(fieldType), inputType)
	case reflect.Bool:
		jsonType = "boolean"
	case reflect.Slice, reflect.Array:
		elementType := fieldType.Elem()
		if elementType.Kind() == reflect.Struct {
			return sc.Build(reflect.New(fieldType).Elem().Interface(), inputType)
		}

	default:
		jsonType = "object"
	}

	schema := map[string]interface{}{
		"type": jsonType,
	}

	if enums := field.Tag.Get("enums"); len(enums) > 0 {
		enumValues := strings.Split(strings.TrimSpace(enums), ",")
		schema["enum"] = enumValues
	}

	return schema
}

func (sc *Schema) buildSchemaForObject() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
	}
}
