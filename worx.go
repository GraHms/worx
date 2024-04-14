package worx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grahms/worx/router"
	"html/template"
	"net/http"
	"time"
)

type Application struct {
	name        string
	path        string
	router      *gin.RouterGroup
	engine      *gin.Engine
	version     string
	description string
}

func NewRouter[In, Out any](app *Application, path string) *router.APIEndpoint[In, Out] {

	return router.New[In, Out](path, app.router.Group(""))
}

func NewApplication(path, name, version, description string, middlewares ...gin.HandlerFunc) *Application {
	r := Engine()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	config.ExposeHeaders = []string{"Content-Length", "X-Result-Count", "X-Total-Count", "Content-Type"}
	r.Use(cors.New(config))
	// Optionally apply custom middleware
	for _, mw := range middlewares {
		r.Use(mw)
	}
	r.HandleMethodNotAllowed = true
	noMethod(r)
	noRoute(r)
	g := r.Group(path)
	app := &Application{
		name:        name,
		path:        path,
		router:      g,
		engine:      r,
		version:     version,
		description: description,
	}
	return app
}

type _ any

func (a *Application) Run(address string) error {
	a.renderDocs()
	return a.engine.Run(address)
}

func (a *Application) renderDocs() {
	s, err := router.NewOpenAPI(a.name, a.version, a.description).SetEndpoints(router.Endpoints).Build()
	if err != nil {
		panic(err)
	}
	bJ, _ := json.Marshal(s)

	a.engine.GET("/spec", RenderSwagg(string(bJ))) // Serve swagger ui
	a.engine.GET("/openapi.json", func(c *gin.Context) {

		c.Header("Content-Type", "application/json")
		c.String(200, string(bJ))
	})
	a.engine.GET("", func(c *gin.Context) {

		c.Data(200, "text/html; charset=utf-8", []byte(redocHTML))
	})

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

func RenderSwagg(spec string) func(c *gin.Context) {
	return func(c *gin.Context) {

		tmplData := struct {
			SwaggerJSON string
		}{
			SwaggerJSON: spec,
		}

		t := template.Must(template.New("swagger").Parse(swagTempl))

		var buf bytes.Buffer
		if err := t.Execute(&buf, tmplData); err != nil {
			c.String(http.StatusInternalServerError, "Failed to render Swagger UI")
			return
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, buf.String())
	}

}

const swagTempl = `
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
const redocHTML = `
   <!DOCTYPE html>
<html>
  <head>
    <title>Redoc</title>
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <!--
    Redoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='/openapi.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
  </body>
</html>

    `
