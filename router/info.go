package router

import (
	"regexp"
	"strings"
)

// Endpoint represents information about an API endpoint
type Endpoint struct {
	Path    string
	Methods []Method
}

type Method struct {
	HTTPMethod  string
	Request     interface{}
	Response    interface{}
	Description string
	StatusCode  *int
	Tags        []string
	Summery     string
	Configs     EndpointConfigs
}

type EndpointConfigs struct {
	Name           string
	Uri            string
	StatusCode     *int
	Tags           []string
	GeneratedTags  []string
	Descriptions   string
	AllowedHeaders []AllowedFields
	AllowedParams  []AllowedFields
	PathParams     []AllowedFields
}

type AllowedFields struct {
	Name        string
	Description string
	Required    bool
}

func getConfigs(opts ...HandleOption) *EndpointConfigs {
	config := &EndpointConfigs{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

var Endpoints map[string]*Endpoint

// Initialize the Endpoints map
func init() {
	Endpoints = make(map[string]*Endpoint)
}

func registerEndpoint(path, method string, request, response interface{}, config EndpointConfigs, opts ...HandleOption) {
	opts = analyzePathParameters(path, opts...)
	re := regexp.MustCompile(`/:(\w+)(/|$)`)
	path = re.ReplaceAllString(path, "/{$1}$2")
	re = regexp.MustCompile(`{(\w+)}(/|$)`)
	_ = re.FindAllStringSubmatch(path, -1)
	config = *getConfigs(opts...)

	if endpoint, ok := Endpoints[path]; ok {
		for _, m := range endpoint.Methods {
			if m.HTTPMethod == method {
				return
			}
		}
		// Method does not exist, add it to the endpoint's methods
		endpoint.Methods = append(endpoint.Methods, Method{
			HTTPMethod:  method,
			Request:     request,
			Response:    response,
			Description: config.Descriptions,
			Tags:        config.Tags,
			Summery:     config.Name,
			Configs:     config,
		})
	} else {
		// If the endpoint doesn't exist, create a new entry with the method
		Endpoints[path] = &Endpoint{
			Path: path,
			Methods: []Method{
				{
					HTTPMethod:  method,
					Request:     request,
					Response:    response,
					Description: config.Descriptions,
					Tags:        config.Tags,
					Summery:     config.Name,
					Configs:     config,
				},
			},
		}
	}
}

type HandleOption func(*EndpointConfigs)

func WithName(name string) HandleOption {
	return func(c *EndpointConfigs) {
		c.Name = name
	}
}

func WithStatusCode(statusCode int) HandleOption {
	return func(c *EndpointConfigs) {
		c.StatusCode = &statusCode
	}
}

func WithTags(tags []string) HandleOption {
	return func(c *EndpointConfigs) {
		c.Tags = append(tags)
	}
}

func WithDescriptions(descriptions string) HandleOption {
	return func(c *EndpointConfigs) {
		c.Descriptions = descriptions
	}
}

func WithAllowedHeaders(headers []AllowedFields) HandleOption {
	return func(c *EndpointConfigs) {
		c.AllowedHeaders = headers
	}
}

func WithAllowedParams(params []AllowedFields) HandleOption {
	return func(c *EndpointConfigs) {
		c.AllowedParams = params
	}
}
func withPathParams(params []AllowedFields) HandleOption {
	return func(c *EndpointConfigs) {
		c.PathParams = params
	}
}

func analyzePathParameters(path string, opts ...HandleOption) []HandleOption {
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			paramName := strings.TrimPrefix(segment, ":")
			pathParamConfig := []AllowedFields{
				{
					Name:     paramName,
					Required: true,
				},
			}
			opts = append(opts, withPathParams(pathParamConfig))
		}
	}
	return opts
}
