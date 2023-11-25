package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grahms/godantic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Req struct {
	Foo *string `json:"foo"`
	Baz *int    `json:"baz"`
}

type Resp struct {
	Foo string `json:"foo"`
	Baz int    `json:"baz"`
}

func TestCreate(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Req]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Req) {
		foo := "bar"
		return nil, &Req{
			Foo: &foo,
			Baz: nil,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Unmarshal the response body into a map
	var respMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respMap); err != nil {
		t.Fatal(err)
	}

	// Check that the response body is correct
	assert.Equal(t, map[string]interface{}{
		"foo": "bar",
	}, respMap)
}

func TestShouldUpdate(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(id string, req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleUpdate("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPatch, "/foo/id", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Unmarshal the response body into a map
	var respMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respMap); err != nil {
		t.Fatal(err)
	}

	// Check that the response body is correct
	assert.Equal(t, map[string]interface{}{
		"foo": "bar",
		"baz": float64(123),
	}, respMap)
}

func TestShouldUpdateWithWrongContentType(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(id string, req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleUpdate("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPatch, "/foo/id", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("foo", "bar")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

}

func TestShouldUpdateWithWrongJson(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(id string, req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleUpdate("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPatch, "/foo/id", bytes.NewBuffer([]byte(`"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestShouldNotUpdateWithProccesorErr(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(id string, req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return &ProcessorError{
			StatusCode: 500,
		}, nil
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleUpdate("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPatch, "/foo/id", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

}

func TestShouldTryCreateWithWrongJsonFormat(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo",
		bytes.NewBuffer([]byte(`{"foo":"bar","baz":123`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("foo", "bar")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestShouldTryCreateWithEmptyFIeld(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo",
		bytes.NewBuffer([]byte(`{"foo":"","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestShouldTryCreateWithExtraField(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo",
		bytes.NewBuffer([]byte(`{"foo":"bar","ismael":"grahms","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestShouldTryCreateWithRequestProcessorError(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		perr := &ProcessorError{
			err:        errors.New("i'm an error"),
			StatusCode: 500,
			ErrCode:    "ERR_CODE",
			ErrReason:  "REASON",
		}
		return perr, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo",
		bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

}

func TestCreateWithInvalidBody(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer([]byte(`{"fo}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestCreateWithEmptyBody(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatal(err)
	}
	// Set the Content-Type header of the request
	req.Header.Set("Content-Type", "application/json")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}
func TestCreateWithWrongContentType(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(req Req, _ *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleCreate("", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodPost, "/foo", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

}

func TestRead(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(values *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleRead("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request with headers
	req, err := http.NewRequest(http.MethodGet, "/foo/id?fields=foo,baz", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("foo", "bar")

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Unmarshal the response body into a map
	var respMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respMap); err != nil {
		t.Fatal(err)
	}

	// Check that the response body is correct
	assert.Equal(t, map[string]interface{}{
		"foo": "bar",
		"baz": float64(123),
	}, respMap)
}

func TestReadWithoutFields(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(values *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleRead("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodGet, "/foo/id", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Unmarshal the response body into a map
	var respMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respMap); err != nil {
		t.Fatal(err)
	}

	// Check that the response body is correct
	assert.Equal(t, map[string]interface{}{
		"foo": "bar",
		"baz": float64(123),
	}, respMap)
}

func TestReadWithoutProcessorError(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(values *RequestParams) (*ProcessorError, *Resp) {
		return &ProcessorError{
			StatusCode: 500,
			ErrCode:    "someErrCode",
			ErrReason:  "NoReason",
			Message:    "Wassap error",
		}, nil
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleRead("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodGet, "/foo/id", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Unmarshal the response body into a map
	var respMap map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respMap); err != nil {
		t.Fatal(err)
	}

	// Check that the response body is correct
	assert.Equal(t, map[string]interface{}{"code": "someErrCode", "message": "Wassap error", "reason": "NoReason"}, respMap)
}

func TestReadWithInvalidFIelds(t *testing.T) {
	// HandleCreate a new APIEndpoint instance
	engine := gin.New()
	router := engine.Group("")
	e := &APIEndpoint[Req, Resp]{
		Path:       "/foo",
		Router:     router,
		validator:  &Validation{},
		dataBinder: godantic.Validate{},
	}

	// Define a request processor function that returns a response body and a nil error
	requestProcessor := func(values *RequestParams) (*ProcessorError, *Resp) {
		return nil, &Resp{
			Foo: "bar",
			Baz: 123,
		}
	}

	// Register the request processor function with the APIEndpoint instance
	e.HandleRead("/:id", requestProcessor)

	// HandleCreate a mock HTTP POST request
	req, err := http.NewRequest(http.MethodGet, "/foo/id?fields=foo,baz,grahms", bytes.NewBuffer([]byte(`{"foo":"bar","baz":123}`)))
	if err != nil {
		t.Fatal(err)
	}

	// HandleCreate a response recorder
	rr := httptest.NewRecorder()

	// Serve the mock HTTP request
	engine.ServeHTTP(rr, req)

	// Check that the response has a 200 status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

}

//func removeNilPointers(data map[string]interface{}) {
//	for key, value := range data {
//		if value == nil {
//			delete(data, key)
//			continue
//		}
//
//		switch v := value.(type) {
//
//		case map[string]interface{}:
//			removeNilPointers(v)
//
//			if len(v) == 0 {
//				delete(data, key)
//			}
//		case []interface{}:
//			removeNilPointersFromArray(v)
//			if len(v) == 0 {
//				delete(data, key)
//			}
//		}
//	}
//}
//
//func removeNilPointersFromArray(arr []interface{}) {
//	for i := 0; i < len(arr); i++ {
//		if arr[i] == nil {
//			arr = append(arr[:i], arr[i+1:]...)
//			i--
//		} else if m, ok := arr[i].(map[string]interface{}); ok {
//			removeNilPointers(m)
//			if len(m) == 0 {
//				arr = append(arr[:i], arr[i+1:]...)
//				i--
//			}
//		} else if subArr, ok := arr[i].([]interface{}); ok {
//			removeNilPointersFromArray(subArr)
//			if len(subArr) == 0 {
//				arr = append(arr[:i], arr[i+1:]...)
//				i--
//			}
//		}
//	}
//}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked read error")
}

func TestBindJSON_ReadError(t *testing.T) {
	r := &APIEndpoint[Req, Resp]{} // assuming APIEndpoint has been defined elsewhere

	var v any // 'any' is a placeholder; adjust as per your actual type
	err := r.bindJSON(&errorReader{}, &v)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if err.Error() != "mocked read error" {
		t.Fatalf("expected error: mocked read error, got: %s", err.Error())
	}
}
