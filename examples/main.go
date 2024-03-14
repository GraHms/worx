package main

import (
	"github.com/grahms/worx"
	"github.com/grahms/worx/router"
)

type Address struct {
	Name *string `json:"name"`
}

type Spec struct {
	Name   *string     `json:"name"`
	Value  *string     `json:"value"`
	Adress *[]*Address `json:"adress"`
}

type Product struct {
	Name          *string  `json:"name" binding:"required"`
	Price         *float64 `json:"price"`
	BaseType      *string  `json:"@baseType"  binding:"ignore"`
	Type          *string  `json:"type" enums:"physical,digital"`
	Url           *string  `json:"@Url"  binding:"ignore"`
	Specification *[]Spec  `json:"specification"`
	Id            *string  `json:"id"`
}

func main() {
	app := worx.NewApplication("/api", "Product Catalog API")
	product := worx.NewRouter[Product, Product](app, "/products")
	// scope, role  =
	product.HandleCreate("",
		func(product Product, params *router.RequestParams) (*router.Err, *Product) {
			return nil, &product
		})
	product.HandleRead("", handler)
	//product.HandleRead("/sidy", handler)
	err := app.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func handler(params *router.RequestParams) (*router.Err, *Product) {
	return nil, &Product{}
}
