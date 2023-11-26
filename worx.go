package worx

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grahms/worx/router"
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
	r := gin.Default()

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
