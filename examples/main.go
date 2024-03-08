package main

import (
	"github.com/grahms/worx"
	"github.com/grahms/worx/router"
)

type Product struct {
	Name     *string  `json:"name" binding:"required"`
	Price    *float64 `json:"price"`
	BaseType *string  `json:"@baseType"`
	Url      *string  `json:"@Url"`
	ID       string   `json:"id" binding:"ignore"`
}

func main() {
	app := worx.NewApplication("/api", "Product Catalog API")
	product := worx.NewRouter[Product, Product](app, "/products")
	product.HandleCreate("",
		func(product Product, params *router.RequestParams) (*router.ProcessorError, *Product) {
			return nil, &product
		})

	err := app.Run(":8080")
	if err != nil {
		panic(err)
	}
}
