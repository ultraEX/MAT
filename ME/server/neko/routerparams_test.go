package neko

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_RouterParams(t *testing.T) {
	m := New()
	m.GET("/params/:name", func(ctx *Context) {
		So(ctx.Params.ByGet("name"), ShouldEqual, "neko")
		So(ctx.Params.ByGet("say"), ShouldEqual, "hello")
	})
	m.POST("/params/:name", func(ctx *Context) {
		So(ctx.Params.ByGet("name"), ShouldEqual, "neko")
		So(ctx.Params.ByPost("say"), ShouldEqual, "hello")
	})

	m.POST("/json/:name", func(ctx *Context) {
		dataJson := ctx.Params.Json()
		So(ctx.Params.ByGet("name"), ShouldEqual, "neko")

		So(dataJson.Get("say"), ShouldEqual, "hello")
		So(dataJson.GetString("say"), ShouldEqual, "hello")
		So(dataJson.GetInt32("int32"), ShouldEqual, 1)
		So(dataJson.GetUInt32("uint32"), ShouldEqual, 2)
		So(dataJson.GetFloat32("float32"), ShouldEqual, 3)
		So(dataJson.GetFloat64("float64"), ShouldEqual, 4)

		So(dataJson.String(), ShouldEqual, `{"say": "hello", "int32": "1", "uint32": "2", "float32": "3", "float64": "4"}`)
	})

	m.POST("/json-empty", func(ctx *Context) {
		dataJson := ctx.Params.Json()
		So(dataJson.String(), ShouldEqual, "")
		So(dataJson.Get("empty"), ShouldEqual, "")
		So(dataJson.GetString("say"), ShouldEqual, "")
		So(dataJson.GetInt32("int32"), ShouldEqual, 0)
		So(dataJson.GetUInt32("uint32"), ShouldEqual, 0)
		So(dataJson.GetFloat32("float32"), ShouldEqual, 0)
		So(dataJson.GetFloat64("float64"), ShouldEqual, 0)
	})

	m.POST("/bindjson", func(ctx *Context) {
		type mockList struct {
			Label string
			Num   int
		}
		type mockBindJSON struct {
			Name string
			List []mockList
		}

		var mock mockBindJSON
		err := ctx.Params.BindJSON(&mock)
		So(err, ShouldBeNil)
		So(mock.Name, ShouldEqual, "neko")
		So(len(mock.List), ShouldEqual, 2)
		So(mock.List[0].Label, ShouldEqual, "item-1")
		So(mock.List[0].Num, ShouldEqual, 1)
		So(mock.List[1].Label, ShouldEqual, "item-2")
		So(mock.List[1].Num, ShouldEqual, 2)
	})

	Convey("Get Params By Query String", t, func() {
		performRequest(m, "GET", "/params/neko?say=hello&name=golang", "")
	})
	Convey("Get Params By Form Post", t, func() {
		performRequest(m, "POST", "/params/neko", "say=hello&name=golang")
	})
	Convey("Get Params By Json Data", t, func() {
		performRequest(m, "POST|JSON", "/json/neko", `{"say": "hello", "int32": "1", "uint32": "2", "float32": "3", "float64": "4"}`)
		performRequest(m, "POST|JSON", "/json-empty", "")
	})
	Convey("Get Params By `BindJSON`", t, func() {
		performRequest(m, "POST|JSON", "/bindjson", `{"name": "neko", "list": [{"label": "item-1", "num": 1}, {"label": "item-2", "num": 2}]}`)
	})
}
