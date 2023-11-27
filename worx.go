package worx

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grahms/worx/router"
	"time"
)

type Application struct {
	name   string
	path   string
	router *gin.RouterGroup
	engine *gin.Engine
}

func NewRouter[In, Out any](path string) *router.APIEndpoint[In, Out] {
	return router.New[In, Out](path)
}

func NewApplication(path, name string) *Application {
	r := Engine()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	config.ExposeHeaders = []string{"Content-Length", "X-Result-Count", "X-Total-Count", "Content-Type"}
	r.Use(cors.New(config))
	r.HandleMethodNotAllowed = true
	noMethod(r)
	noRoute(r)
	g := r.Group(path)
	return &Application{
		name:   name,
		path:   path,
		router: g,
		engine: r,
	}
}

type _ any

func IncludeRoute[In, Out any](a *Application, apiEndpoint *router.APIEndpoint[In, Out]) {
	apiEndpoint.Router = a.router.Group("")
}

func (a *Application) Run(address string) error {
	return a.engine.Run(address)
}
func noRoute(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		var e *router.Error
		c.JSON(e.ResourceNotFound())
		return
	})
}

func noMethod(r *gin.Engine) {
	r.NoMethod(func(c *gin.Context) {
		var e *router.Error
		c.JSON(e.MethodNotAllowed())
		return
	})

}

func Engine() *gin.Engine {

	engine := gin.New()
	engine.Use(Logger(), gin.Recovery())
	return engine
}

func Logger() gin.HandlerFunc {
	return LoggerWithConfig(gin.LoggerConfig{})
}

func LoggerWithConfig(conf gin.LoggerConfig) gin.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter
	}

	notlogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			fmt.Fprint(out, formatter(param))
		}
	}
}

// defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	logData := map[string]interface{}{
		"timestamp": param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		"status":    param.StatusCode,
		"latency":   param.Latency.String(),
		"client_ip": param.ClientIP,
		"method":    param.Method,
		"path":      param.Path,
		"error":     param.ErrorMessage,
	}

	logJSON, err := json.Marshal(logData)
	if err != nil {
		// Handle error, log or return a default value
		return fmt.Sprintf("[WORX] Error formatting log as JSON: %v\n", err)
	}
	return string(logJSON) + "\n"
}
