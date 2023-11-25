package main

import (
	"github.com/grahms/worx"
	"github.com/grahms/worx/router"
)

type Product struct {
	Name  *string  `json:"name"`
	Price *float64 `json:"price"`
}
type ProductResponse struct {
	Product
	BaseType *string `json:"@baseType"`
	Url      *string `json:"@Url"`
}

func main() {
	app := worx.NewApplication("/api", "Product Catalog API")
	product := worx.NewRouter[Product, ProductResponse]("/products", Product{}, ProductResponse{})
	worx.IncludeRoute(app, product)
	product.HandleCreate("", func(product Product, params *router.RequestParams) (*router.ProcessorError, *ProductResponse) {
		return nil, &ProductResponse{
			Product:  product,
			BaseType: nil,
			Url:      nil,
		}
	})

	err := app.Run(":8080")
	if err != nil {
		panic(err)
	}
}
