

# Getting Started

Worx is a framework for building APIs in Go with support for TMF (Telecom Management Forum) standards. Follow these steps to get started:

### 1. Install Worx:

```bash
go get -u github.com/grahms/worx
```

### 2. Initialize Your Application:

Create a new Worx application:

```go
package main

import (
	"github.com/grahms/worx"
	"github.com/grahms/worx/router"
)

func main() {
	app := worx.NewApplication("/api", "Product Catalog API")
}
```

### 3. Define Your API Endpoint:

Create a new API endpoint using the `NewRouter` function:

```go
product := worx.NewRouter[Product, ProductResponse]("/products", Product{}, ProductResponse{})
```

### 4. Include Routes in Your Application:

Include the defined route in your Worx application **before defining handlers**:

```go
worx.IncludeRoute(app, product)
```

### 5. Handle Requests:

Define your request handling logic using the `HandleCreate`, `HandleRead`, `HandleUpdate`, and `HandleList` methods:

```go
product.HandleCreate("", func(product Product, params *router.RequestParams) (*router.ProcessorError, *ProductResponse) {
	return nil, &ProductResponse{
		Product:  product,
		BaseType: nil,
		Url:      nil,
	}
})
```

### 6. Run Your Application:

Start your Worx application and listen on a specified port:

```go
err := app.Run(":8080")
if err != nil {
	panic(err)
}
```

Now, your Worx application is ready to handle TMF API requests.

--- 

Make sure to include the route before defining the handlers to ensure that the routes are properly registered in your application. Happy coding