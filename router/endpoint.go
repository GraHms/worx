package router

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/grahms/godantic"
	"io"

	"net/http"
	"strconv"
	"strings"
)

type Handler interface {
	Handler()
}

func NewAPIEndpointGroup[Req any, Resp any](path string, group *gin.RouterGroup) *APIEndpoint[Req, Resp] {
	return &APIEndpoint[Req, Resp]{
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
	TraceID    string
}

func New[In, Out any](path string, group *gin.RouterGroup) *APIEndpoint[In, Out] {
	return &APIEndpoint[In, Out]{
		Path:   path,
		Router: group,
	}
}
func setTags(path string, config *EndpointConfigs) {
	tag := strings.TrimPrefix(path, "/")
	tag = strings.ReplaceAll(path, "/", " ")
	tag = strings.TrimPrefix(tag, " ")
	config.GeneratedTags = []string{tag}
}
func (r *APIEndpoint[Req, Resp]) HandleCreate(uri string, processRequest func(Req, *RequestParams) (*Err, *Resp), opts ...HandleOption) {
	config := getConfigs(opts...)
	setTags(r.Path, config)
	registerEndpoint(r.Router.BasePath()+r.Path+uri, "POST", new(Req), new(Resp), *config, opts...)
	statusCode := http.StatusCreated
	if config.StatusCode != nil {
		statusCode = *config.StatusCode
	}
	r.Router.POST(r.Path+uri, func(c *gin.Context) {

		params := r.extractRequestParams(c)
		if statusCode, exception := r.validateJSONContentType(c); exception != nil {

			c.JSON(statusCode, exception)
			return
		}

		var requestBody Req
		if err := r.bindJSON(c.Request.Body, &requestBody); err != nil {
			code, e := r.validator.InputErr(err)

			c.JSON(code, e)
			return
		}

		perr, response := processRequest(requestBody, &params)
		if perr != nil {
			code, e := r.validator.ProcessorErr(perr)

			c.JSON(code, e)
			return
		}

		c.JSON(statusCode, r.convertToMap(*response))
		return
	})
}

func (r *APIEndpoint[Req, Resp]) HandleRead(pathSuffix string, processRequest func(*RequestParams) (*Err, *Resp), opts ...HandleOption) {
	config := getConfigs(opts...)
	setTags(r.Path, config)
	registerEndpoint(r.Router.BasePath()+r.Path+pathSuffix, "GET", new(Req), new(Resp), *config, opts...)
	statusCode := http.StatusOK
	if config.StatusCode != nil {
		statusCode = *config.StatusCode
	}
	r.Router.GET(r.Path+pathSuffix, func(c *gin.Context) {
		reqValues := r.extractRequestParams(c)
		perr, resp := processRequest(&reqValues)
		// handle processor error
		if perr != nil {
			code, e := r.validator.ProcessorErr(perr)

			c.JSON(code, e)
			return
		}
		fields := strings.Replace(c.Query("fields"), " ", "", -1)
		fieldsList := make([]string, 0)
		if fields != "" {
			fieldsList = strings.Split(fields, ",")

		}
		respWithFields, err := fieldSelector(fieldsList, r.convertToMap(*resp))
		if err != nil {
			code, e := r.validator.ProcessorErr(err)

			c.JSON(code, e)
			return
		}

		c.JSON(statusCode, respWithFields)
		return
	})
}

func (r *APIEndpoint[Req, Resp]) HandleUpdate(pathString string, requestProcessor func(id string, reqBody Req, params *RequestParams) (*Err, *Resp), opts ...HandleOption) {
	binder := godantic.Validate{}
	binder.IgnoreRequired = true
	binder.IgnoreMinLen = true

	config := getConfigs(opts...)
	setTags(r.Path, config)
	registerEndpoint(r.Router.BasePath()+r.Path+pathString, "PATCH", new(Req), new(Resp), *config, opts...)
	statusCode := http.StatusOK

	r.Router.PATCH(r.Path+pathString, func(c *gin.Context) {
		if statusCode, exception := r.validateJSONContentType(c); exception != nil {
			c.JSON(statusCode, exception)
			return
		}

		var reqBody Req
		reqValues := r.extractRequestParams(c)
		requestDataBytes, err := io.ReadAll(c.Request.Body)
		if err = binder.BindJSON(requestDataBytes, &reqBody); err != nil {
			code, e := r.validator.InputErr(err)

			c.JSON(code, e)
			return
		}
		id := c.Param("id")

		perr, resp := requestProcessor(id, reqBody, &reqValues)
		if perr != nil {
			code, e := r.validator.ProcessorErr(perr)

			c.JSON(code, e)
			return
		}

		c.JSON(statusCode, r.convertToMap(*resp))
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

func (r *APIEndpoint[Req, Resp]) HandleList(pathString string, requestProcessor func(params *RequestParams, limit int, offset int) ([]*Resp, *Err, int, int), opts ...HandleOption) {
	opts = append(opts, WithAllowedParams([]AllowedFields{
		{
			Name:        "limit",
			Description: "page limit",
		},
		{
			Name:        "offset",
			Description: "page number",
		},
		{
			Name:        "fields",
			Description: "fields to be selected ex: fields=id,name",
		},
	}))
	config := getConfigs(opts...)
	setTags(r.Path, config)

	registerEndpoint(r.Router.BasePath()+r.Path+pathString, "GET", new(Req), new(Resp), *config)
	statusCode := http.StatusOK
	if config.StatusCode != nil {
		statusCode = *config.StatusCode
	}
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
			code, e := r.validator.ProcessorErr(perr)

			c.JSON(code, e)
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
				code, e := r.validator.ProcessorErr(err)
				c.JSON(code, e)
				return
			}
			responseMaps = append(responseMaps, respWithFields)
		}

		c.Header("X-Total-Count", strconv.Itoa(amount))
		c.Header("X-Result-Count", strconv.Itoa(total))
		c.Header("trace-id", params.TraceID)
		if len(resp) == 0 {
			c.JSON(http.StatusOK, resp)
			return
		}
		c.JSON(statusCode, responseMaps)
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

func (r *APIEndpoint[Req, Resp]) HandleCreateWithoutBody(uri string, processRequest func(*RequestParams) (*Err, *Resp), opts ...HandleOption) {
	config := getConfigs(opts...)
	setTags(r.Path, config)
	registerEndpoint(r.Router.BasePath()+r.Path+uri, "POST", new(Req), new(Resp), *config)
	statusCode := http.StatusCreated
	if config.StatusCode != nil {
		statusCode = *config.StatusCode
	}
	r.Router.POST(r.Path+uri, func(c *gin.Context) {
		params := r.extractRequestParams(c)
		perr, response := processRequest(&params)
		if perr != nil {
			c.JSON(r.validator.ProcessorErr(perr))
			return
		}

		c.JSON(statusCode, r.convertToMap(*response))
		return
	})
}

func (r *APIEndpoint[Req, Resp]) HandleDelete(pathString string, processRequest func(params *RequestParams) *Err, opts ...HandleOption) {
	config := getConfigs(opts...)
	setTags(r.Path, config)
	registerEndpoint(r.Router.BasePath()+r.Path+pathString, "DELETE", nil, nil, *config, opts...)
	statusCode := http.StatusNoContent
	config.StatusCode = &statusCode
	r.Router.DELETE(r.Path+pathString, func(c *gin.Context) {
		params := r.extractRequestParams(c)
		perr := processRequest(&params)
		if perr != nil {
			code, e := r.validator.ProcessorErr(perr)

			c.JSON(code, e)
			return
		}

		c.Status(statusCode)
		return
	})
}

func WithContext(params *RequestParams) context.Context {
	return context.WithValue(context.Background(), "traceID", params.TraceID)
}
