#Neko
[![wercker status](https://app.wercker.com/status/2ab4b79cf2d418606e884c5d98d1ec0d/s "wercker status")](https://app.wercker.com/project/bykey/2ab4b79cf2d418606e884c5d98d1ec0d)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/rocwong/neko)
[![GoCover](http://gocover.io/_badge/github.com/rocwong/neko)](http://gocover.io/github.com/rocwong/neko)



A lightweight web application framework for Golang

**NOTE: Neko is still under development, so API might be changed in future.**

## Features

* Extremely simple to use.
* RESTful support
* Middleware support
* Unlimited nested group routers.

## Getting Started
Basic usage
~~~go
package main
import "github.com/rocwong/neko"
func main() {
  app := neko.Classic()
  app.GET("/", func(ctx *neko.Context)  {
      ctx.Text("Hello world!")
  })
  app.Run(":3000")
}
~~~
Initial Neko without middlewares
~~~go
app := neko.New()
app.Use(neko.Logger())
app.Use(neko.Recovery())
~~~

##Routing
Using GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS and Any
~~~go
app.GET("/get", get)
app.POST("/post", post)
app.PUT("/put", put)
app.PATCH("/patch", patch)
app.DELETE("/delete", delete)
app.HEAD("/head", head)
app.OPTIONS("/options", options)
// A shortcut for all request methods
app.Any("/any", options)
~~~
Neko uses julienschmidt's [httprouter](https://github.com/julienschmidt/httprouter) internaly.


##Group Routing
~~~go
v1 := app.Group("/v1", func(router *neko.RouterGroup) {
  // Match /v1/item
  router.GET("/item", item)

  // Nested group
  router.Group("/sub", func(sub *neko.RouterGroup) {
    // Match /v1/sub/myitem
    sub.GET("/myitem", myitem)
  })
})
// Match /v1/act
v1.GET("/act", act)
~~~

## Parameters

####Parameters in path
~~~go
app.GET("/user/:name/*age", func(ctx *neko.Context) {
  // Request: "/user/neko/1?name=none&food=fish"
  
  name := ctx.Params.ByGet("name")
  age := ctx.Params.ByGet("age")
  food := ctx.Params.ByGet("food")
  
  // Response: neko is 1, eat fish
  ctx.Text(name + " is " + age + ", eat " + ctx.Params.ByGet("eat"))
})
~~~

####Multipart/Urlencoded Form
~~~go
app.POST("/user", func(ctx *neko.Context) {
  // Request: "/user"  Post Data: { name: neko, age: 1}
  
  // Response: neko is 1
  ctx.Text(ctx.Params.ByPost("name") + " is " + ctx.Params.ByPost("age"))
})
~~~

####Json Data
~~~go
app.POST("/user", func(ctx *neko.Context) {
  // Request: "/user"  Post Data: { name: neko, age: 1} Content-type: "application/json"
  
  // Only get once
  dataJson := ctx.Params.Json()
  
  // Response: neko is 1
  ctx.Text(dataJson.GetString("name") + " is " + dataJson.Get("age"))
  
  // dataJson.Get(param) : return type interface {}
  // dataJson.GetString(param) : return type string
  // dataJson.GetInt32(param) : return type int32
  // dataJson.GetUInt32(param) : return type uint32
  // dataJson.GetFloat32(param) : return type float32
  // dataJson.GetFloat64(param) : return type float64
})
~~~

####BindJSON
~~~go
type User struct {
  User     string
  Password string
}
app.POST("/user", func(ctx *neko.Context) {
  // Request: "/user"  Post Data: { name: neko, password: abcdefg} Content-type: "application/json"
  var json User
  if ctx.Params.BindJSON(&json) == nil {
    // Response: neko's password abcdefg
    ctx.Text(json.Name + "'s password " + json.Password)
  }
})
~~~

##Response

####Render
~~~go
type ExampleXml struct {
  XMLName xml.Name `xml:"example"`
  One     string   `xml:"one,attr"`
  Two     string   `xml:"two,attr"`
}

// Response: <example one="hello" two="xml"/>
ctx.Xml(ExampleXml{One: "hello", Two: "xml"})
~~~

~~~go
// Response: {"msg": "json render", "status": 200}
ctx.Json(neko.JSON{"msg": "json render", "status": 200})

// Response: neko({"msg": "json render", "status": 200})
ctx.Jsonp("neko", neko.JSON{"msg": "json render", "status": 200})

// Response: neko text
ctx.Text("neko text")
~~~

####Redirect
~~~go
// Default 302
ctx.Redirect("/")

// Redirect 301
ctx.Redirect("/", 301)
~~~

####Headers
~~~go
// Get header
ctx.Writer.Header()

// Set header
ctx.SetHeader("x-before", "before")
~~~

##Cookie
~~~ go
app.GET("/", func (ctx *neko.Context) {
  ctx.SetCookie("myvalue", "Cookies Save")
  ctx.Text("Cookies Save")
})

app.GET("/get", func (ctx *neko.Context) {
  ctx.Text(ctx.GetCookie("myvalue"))
})
~~~
####Secure cookie
~~~ go
// Set cookie secret
app.SetCookieSecret("secret123")

app.GET("/set-secure", func (ctx *neko.Context) {
  ctx.SetSecureCookie("sv", "Cookies Save")
  ctx.Text("Cookies Save")
})

app.GET("/get-secure", func (ctx *neko.Context) {
  ctx.Text(ctx.GetSecureCookie("sv"))
})

~~~
Use following arguments order to set more properties:

`SetCookie/SetCookieSecret(name, value [, MaxAge, Path, Domain, Secure, HttpOnly])`.

## Middlewares

####Using middlewares
~~~go
// Global middlewares
app.Use(neko.Logger())

// Per route middlewares, you can add as many as you desire.
app.Get("/user", mymiddleware(), mymiddleware2(), user)

// Pass middlewares to groups
v1 := app.Group("/v1", func(router *neko.RouterGroup) {
  router.GET("/item", item)
}, mymiddleware1(), mymiddleware2(), mymiddleware3())

v1.Use(mymiddleware4)
~~~

####Custom middlewares
~~~go
func mymiddleware() neko.HandlerFunc {
  return func (ctx *neko.Context) {
    // Before request
    t := time.Now()

    ctx.Next()

    // After request
    latency := time.Since(t)
    log.Print(latency)

    // Access the status we are sending
    status := ctx.Writer.Status()
    log.Println(status)
  }
}
~~~

#### More middleware
For more middleware and functionality, check out the repositories in the  [neko-contrib](https://github.com/neko-contrib) organization.

## Others
~~~go
// Static Serves
app.Static("/static", "content/static")

// Get Remote IP Address
app.GET("/", func (ctx *neko.Context) {
  ctx.ClientIP()
}

// Metadata Management
app.GET("/", func (ctx *neko.Context) {
  ctx.Set("foo", "bar")
  v, err := ctx.Get("foo")
  v = ctx.MustGet("foo")
}
~~~


## Credits & Thanks
I use code/got inspiration from these excellent libraries:

*  [Gin](https://github.com/gin-gonic/gin)
*  [Martini](https://github.com/go-martini/martini)
*  [Httprouter](https://github.com/julienschmidt/httprouter)


## License
Neko is licensed under the MIT