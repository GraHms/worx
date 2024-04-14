package main

import (
	"github.com/grahms/worx"
	"github.com/grahms/worx/router"
)

type Address struct {
	Name *string `json:"name" binding:"required"`
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
	Id            *string  `json:"id" binding:"ignore"`
}

func main() {
	app := worx.NewApplication("/api", "Product API", "1.0.0", "Product API")
	productTags := router.WithTags([]string{"Product", "something"})
	product := worx.NewRouter[Product, Product](app, "/products")
	product.HandleCreate("", createHandler, router.WithName("product name"), productTags)
	product.HandleRead("", handler, productTags)
	product.HandleRead("/:id", handler, productTags)
	err := app.Run(":8081")
	if err != nil {
		panic(err)
	}
}

func createHandler(product Product, params *router.RequestParams) (*router.Err, *Product) {
	if *product.Price < 5 {
		return &router.Err{
			StatusCode: 409,
			ErrCode:    "SOME_ERROR",
			ErrReason:  "Price not goood",
			Message:    "The message",
		}, nil
	}

	return nil, &product
}

func handler(params *router.RequestParams) (*router.Err, *Product) {
	return nil, &Product{}
}
