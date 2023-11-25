package worx

import "github.com/gin-gonic/gin"

type Application struct {
	Name   string
	Path   string
	Router *gin.RouterGroup
}

func NewApplication(path, name string) *Application {
	return &Application{
		Name:   name,
		Path:   path,
		Router: nil,
	}
}
