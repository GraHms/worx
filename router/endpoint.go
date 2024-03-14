package router

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/grahms/godantic"
	"io"
	"net/http"
	"strconv"
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
}

var Endpoints map[string]*Endpoint

// Initialize the Endpoints map
func init() {
	Endpoints = make(map[string]*Endpoint)
}

// we can have the global struct here
type APIEndpointer[Req, Resp any] interface {
	HandleCreate(uri string, processRequest func(Req, *RequestParams) (*Err, *Resp))
	HandleRead(pathSuffix string, processRequest func(*RequestParams) (*Err, *Resp))
	HandleUpdate(pathString string, requestProcessor func(id string, reqBody Req, params *RequestParams) (*Err, *Resp))
	HandleList(pathString string, requestProcessor func(params *RequestParams, limit int, offset int) ([]*Resp, *Err, int, int))
}

func New[In, Out any](path string, group *gin.RouterGroup) *APIEndpoint[In, Out] {
	return &APIEndpoint[In, Out]{
		Path:   path,
		Router: group,
	}
}

type APIEndpoint[Req any, Resp any] struct {
	Path       string
	Router     *gin.RouterGroup
	validator  *Validation
	dataBinder godantic.Validate
}

type RequestParams struct {
	Query      map[string]any
	Headers    map[string]string
	PathParams map[string]string
}

func (r *APIEndpoint[Req, Resp]) HandleCreate(uri string, processRequest func(Req, *RequestParams) (*Err, *Resp)) {
	registerEndpoint(r.Router.BasePath()+r.Path+uri, "POST", new(Req), new(Resp))
	r.Router.POST(r.Path+uri, func(c *gin.Context) {
		if statusCode, exception := r.validateJSONContentType(c); exception != nil {
			c.JSON(statusCode, exception)
			return
		}

		var requestBody Req
		if err := r.bindJSON(c.Request.Body, &requestBody); err != nil {
			c.JSON(r.validator.InputErr(err))
			return
		}

		params := r.extractRequestParams(c)
		perr, response := processRequest(requestBody, &params)
		if perr != nil {
			c.JSON(r.validator.ProcessorErr(perr))
			return
		}

		c.JSON(http.StatusCreated, r.convertToMap(*response))
		return
	})
}

func (r *APIEndpoint[Req, Resp]) HandleRead(pathSuffix string, processRequest func(*RequestParams) (*Err, *Resp)) {
	registerEndpoint(r.Router.BasePath()+r.Path+pathSuffix, "GET", new(Req), new(Resp))
	r.Router.GET(r.Path+pathSuffix, func(c *gin.Context) {
		reqValues := r.extractRequestParams(c)
		perr, resp := processRequest(&reqValues)
		// handle processor error
		if perr != nil {
			c.JSON(r.validator.ProcessorErr(perr))
			return
		}
		fields := strings.Replace(c.Query("fields"), " ", "", -1)
		fieldsList := make([]string, 0)
		if fields != "" {
			fieldsList = strings.Split(fields, ",")

		}
		respWithFields, err := fieldSelector(fieldsList, r.convertToMap(*resp))
		if err != nil {
			c.JSON(r.validator.ProcessorErr(err))
			return
		}

		c.JSON(http.StatusOK, respWithFields)
		return
	})
}

func (r *APIEndpoint[Req, Resp]) HandleUpdate(pathString string, requestProcessor func(id string, reqBody Req, params *RequestParams) (*Err, *Resp)) {
	registerEndpoint(r.Router.BasePath()+r.Path+pathString, "PATCH", new(Req), new(Resp))
	binder := godantic.Validate{}
	binder.IgnoreRequired = true
	binder.IgnoreMinLen = true

	r.Router.PATCH(r.Path+pathString, func(c *gin.Context) {
		if statusCode, exception := r.validateJSONContentType(c); exception != nil {
			c.JSON(statusCode, exception)
			return
		}

		var reqBody Req
		requestDataBytes, err := io.ReadAll(c.Request.Body)
		if err = binder.BindJSON(requestDataBytes, &reqBody); err != nil {
			c.JSON(r.validator.InputErr(err))
			return
		}
		id := c.Param("id")
		reqValues := r.extractRequestParams(c)
		perr, resp := requestProcessor(id, reqBody, &reqValues)
		if perr != nil {
			c.JSON(r.validator.ProcessorErr(perr))
			return
		}

		c.JSON(http.StatusOK, r.convertToMap(*resp))
		return
	})
}

func (r *APIEndpoint[Req, Resp]) convertToMap(obj interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	jsonBytes, _ := json.Marshal(obj)
	_ = json.Unmarshal(jsonBytes, &data)
	removeNilPointers(data)
	return data
}

func (r *APIEndpoint[Req, Resp]) HandleList(pathString string, requestProcessor func(params *RequestParams, limit int, offset int) ([]*Resp, *Err, int, int)) {
	registerEndpoint(r.Router.BasePath()+r.Path+pathString, "POST", new(Req), new(Resp))
	r.Router.GET(r.Path+pathString, func(c *gin.Context) {
		params := r.extractRequestParams(c)
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "30"))
		if err != nil {
			c.JSON(http.StatusBadRequest, &Err{
				ErrCode:    "INVALID_LIMIT_ERROR",
				ErrReason:  "Bad Request",
				StatusCode: http.StatusBadRequest,
				Message:    "the query <limit> should be a valid integer",
			})
			return
		}

		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0")) // default offset is 0
		if err != nil {
			c.JSON(http.StatusBadRequest, &Err{
				ErrCode:    "INVALID_OFFSET_ERROR",
				ErrReason:  "Bad Request",
				StatusCode: http.StatusBadRequest,
				Message:    "the query <offset> should be a valid integer",
			})
			return
		}

		resp, perr, amount, total := requestProcessor(&params, limit, offset)
		if perr != nil {
			c.JSON(r.validator.ProcessorErr(perr))
			return
		}

		var responseMaps []map[string]interface{}
		fields := strings.Replace(c.Query("fields"), " ", "", -1)
		fieldsList := make([]string, 0)
		if fields != "" {
			fieldsList = strings.Split(fields, ",")
		}

		for _, res := range resp {
			resp := res
			respMap := r.convertToMap(*resp)
			respWithFields, err := fieldSelector(fieldsList, respMap)
			if err != nil {
				c.JSON(r.validator.ProcessorErr(err))
				return
			}
			responseMaps = append(responseMaps, respWithFields)
		}

		c.Header("X-Total-Count", strconv.Itoa(amount))
		c.Header("X-Result-Count", strconv.Itoa(total))
		if len(resp) == 0 {
			c.JSON(http.StatusOK, resp)
			return
		}
		c.JSON(http.StatusOK, responseMaps)
		return
	})
}

func (r *APIEndpoint[Req, Resp]) validateJSONContentType(c *gin.Context) (int, *Error) {
	if c.ContentType() != "application/json" {
		statusCode, err := r.validator.contentType()
		return statusCode, &err
	}
	return 0, nil
}

func (r *APIEndpoint[Req, Resp]) bindJSON(body io.Reader, v any) error {
	bodyData, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return r.dataBinder.BindJSON(bodyData, v)
}

func (r *APIEndpoint[Req, Resp]) extractRequestParams(c *gin.Context) RequestParams {
	params := RequestParams{
		Query:      make(map[string]any),
		Headers:    make(map[string]string),
		PathParams: make(map[string]string),
	}

	for _, p := range c.Params {
		param := p
		params.PathParams[param.Key] = param.Value
	}
	for k := range c.Request.Header {
		key := k
		params.Headers[key] = c.GetHeader(key)
	}
	for k := range c.Request.URL.Query() {
		key := k
		if key != "limit" && key != "offset" {
			params.Query[key] = c.Query(key)
		}
	}
	return params
}

func GetAuthenticationHeader(req *RequestParams) (string, error) {
	authentication, ok := req.Headers["Authorization"]
	if !ok {
		return "", ErrorBuilder("AUTHORIZATION_HEADER_NOT_FOUND", "Authentication Header Not Found", "Authentication Header Not Found", http.StatusBadRequest)
	}
	return authentication, nil
}
func ErrorBuilder(errCode, errReason, message string, statusCode int) *Err {
	return &Err{
		ErrCode:    errCode,
		ErrReason:  errReason,
		StatusCode: statusCode,
		Message:    message,
	}
}

func registerEndpoint(path, method string, request, response interface{}) {
	// Check if the endpoint already exists
	if endpoint, ok := Endpoints[path]; ok {
		// Check if the method already exists for this endpoint
		for _, m := range endpoint.Methods {
			if m.HTTPMethod == method {
				// Method already exists, return or handle as appropriate
				return
			}
		}
		// Method does not exist, add it to the endpoint's methods
		endpoint.Methods = append(endpoint.Methods, Method{
			HTTPMethod:  method,
			Request:     request,
			Response:    response,
			Description: "",
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
					Description: "",
				},
			},
		}
	}
}
