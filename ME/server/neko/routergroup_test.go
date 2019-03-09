package neko

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func Test_Router(t *testing.T) {
	testRoute("GET", t)
	testRoute("POST", t)
	testRoute("DELETE", t)
	testRoute("PATCH", t)
	testRoute("PUT", t)
	testRoute("OPTIONS", t)
	testRoute("HEAD", t)

	testRouteAny(t)

	testGroup(t)

	testStatic(t)
}

func performRequest(r http.Handler, method, path string, postData string) *httptest.ResponseRecorder {

	req, _ := http.NewRequest(method, path, nil)

	if strings.ToLower(method) == "post" {
		data, _ := url.ParseQuery(postData)
		req, _ = http.NewRequest(method, path, bytes.NewBufferString(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	} else if strings.ToLower(method) == "post|json" {
		req, _ = http.NewRequest("POST", path, bytes.NewBufferString(postData))
		req.Header.Add("Content-Type", "application/json;")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func testRoute(method string, t *testing.T) {
	Convey(method+" Method", t, func() {
		passed := false
		m := New()
		switch method {
		case "GET":
			m.GET("", func(ctx *Context) { passed = true })
		case "POST":
			m.POST("", func(ctx *Context) { passed = true })
		case "DELETE":
			m.DELETE("", func(ctx *Context) { passed = true })
		case "PATCH":
			m.PATCH("", func(ctx *Context) { passed = true })
		case "PUT":
			m.PUT("", func(ctx *Context) { passed = true })
		case "OPTIONS":
			m.OPTIONS("", func(ctx *Context) { passed = true })
		case "HEAD":
			m.HEAD("", func(ctx *Context) { passed = true })
		}
		// RUN
		w := performRequest(m, method, "/", "")

		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func testRouteAny(t *testing.T) {
	Convey("Any Method", t, func() {
		passed := false
		m := New()
		m.Any("", func(ctx *Context) { passed = true })

		w := performRequest(m, "GET", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "POST", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "PUT", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "PATCH", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "HEAD", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "OPTIONS", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "DELETE", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "CONNECT", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "TRACE", "/", "")
		So(passed, ShouldBeTrue)
		So(w.Code, ShouldEqual, http.StatusOK)

		passed = false
		w = performRequest(m, "NO", "/", "")
		So(passed, ShouldBeFalse)
		So(w.Code, ShouldEqual, http.StatusMethodNotAllowed)

	})
}

func testGroup(t *testing.T) {
	Convey("Group routing", t, func() {
		passedGroup, passedGroup2, passedNest := false, false, false
		m := New()
		v1 := m.Group("/v1", func(router *RouterGroup) {
			router.GET("/test", func(ctx *Context) { passedGroup = true })
			router.Group("/sub", func(sub *RouterGroup) {
				sub.GET("/test", func(ctx *Context) { passedNest = true })
			})
		})
		v1.GET("/", func(ctx *Context) { passedGroup2 = true })

		performRequest(m, "GET", "/v1/test", "")
		So(passedGroup, ShouldBeTrue)

		performRequest(m, "GET", "/v1/", "")
		So(passedGroup2, ShouldBeTrue)

		performRequest(m, "GET", "/v1/sub/test", "")
		So(passedNest, ShouldBeTrue)

		w := performRequest(m, "GET", "/v2/test", "")
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}
func testStatic(t *testing.T) {
	Convey("Static serves", t, func() {
		m := New()
		So(func() { m.Static("", "test/") }, ShouldPanic)
		m.Static("/static", "test/")
		w := performRequest(m, "GET", "/static/test.css", "")
		So(w.Code, ShouldEqual, http.StatusOK)
		w = performRequest(m, "GET", "/static/test1.css", "")
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}
