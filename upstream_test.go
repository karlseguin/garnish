package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/garnish/gc"
	"gopkg.in/karlseguin/nd.v1"
	"gopkg.in/karlseguin/typed.v1"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"
)

type UpstreamTests struct{}

func Test_Upstream(t *testing.T) {
	server := startServer()
	defer server.Kill()
	//what could go wrong?
	time.Sleep(time.Second)
	Expectify(new(UpstreamTests), t)
}

func (_ UpstreamTests) Request() {
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Path("/plain").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.Body.String()).To.Equal("hello world")
}

func (_ UpstreamTests) DefaultHeaders() {
	id := nd.LockGuid()
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Host("openmymind.io").Path("/headers").Request)
	Expect(out.Code).To.Equal(200)

	headers, _ := typed.Json(out.Body.Bytes())
	Expect(len(headers)).To.Equal(3)
	Expect(headers.String("x-request-id")).To.Equal(id)
	Expect(headers.String("host")).To.Equal("openmymind.io")
	Expect(headers.String("accept-encoding")).To.Equal("gzip")
}

func (_ UpstreamTests) SpecificHeaders() {
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Host("openmymind.io").Path("/headers").Header("X-Spice", "must flow").Request)
	Expect(out.Code).To.Equal(200)

	headers, _ := typed.Json(out.Body.Bytes())
	Expect(headers.String("x-spice")).To.Equal("must flow")
}

func (_ UpstreamTests) Tweaker() {
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Path("/tweaked").Header("X-Spice", "must flow").Request)
	Expect(out.Code).To.Equal(200)

	headers, _ := typed.Json(out.Body.Bytes())
	Expect(headers.String("x-tweaked")).To.Equal("true")
}

func (_ UpstreamTests) Body() {
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Method("POST").Path("/body").Body("it's over 9000!!").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.Body.String()).To.Equal("it's over 9000!!")
}

func (_ UpstreamTests) DrainedBody() {
	out := httptest.NewRecorder()
	testRuntime().ServeHTTP(out, build.Request().Method("POST").Path("/drain").Body("the spice must flow").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.Body.String()).To.Equal("the spice must flow")
}

func (_ UpstreamTests) ForCache() {
	runtime := testRuntime()
	out := httptest.NewRecorder()
	runtime.ServeHTTP(out, build.Request().Path("/cached").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.Body.String()).To.Equal("will it cache?")

	out = httptest.NewRecorder()
	runtime.ServeHTTP(out, build.Request().Path("/cached").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.Body.String()).To.Equal("will it cache?")
}

func startServer() *os.Process {
	cmd := exec.Command("coffee", "server_test.coffee")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return cmd.Process
}

func testRuntime() *gc.Runtime {
	config := Configure().DnsTTL(-1)
	config.Cache()
	config.Auth(func(req *gc.Request) gc.Response {
		if req.URL.Path == "/drain" {
			if len(req.Body()) == 0 {
				panic("fail")
			}
		}
		return nil
	})
	config.Upstream("test").Address("http://127.0.0.1:4005").KeepAlive(2).Headers("X-Spice")
	config.Upstream("tweaked").Address("http://127.0.0.1:4005").KeepAlive(2).Tweaker(func(in *gc.Request, out *http.Request) {
		out.Header.Set("X-Tweaked", "true")
		out.URL.Path = "/headers"
	})
	config.Upstream("drain").Address("http://127.0.0.1:4005").KeepAlive(2).Tweaker(func(in *gc.Request, out *http.Request) {
		out.URL.Path = "/body"
	})
	config.Route("plain").Get("/plain").Upstream("test")
	config.Route("headers").Get("/headers").Upstream("test")
	config.Route("tweaked").Get("/tweaked").Upstream("tweaked")
	config.Route("body").Post("/body").Upstream("test")
	config.Route("drain").Post("/drain").Upstream("drain")
	config.Route("cached").Get("/cached").Upstream("test").CacheTTL(time.Minute)
	return config.Build()
}
