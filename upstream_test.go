package garnish

import (
	"bytes"
	"encoding/json"
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"github.com/karlseguin/nd"
	"github.com/karlseguin/typed"
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

func (h *UpstreamTests) Request() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/plain").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("11")
	Expect(out.Body.String()).To.Equal("hello world")
}

func (h *UpstreamTests) DefaultHeaders() {
	id := nd.LockGuid()
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Host("openmymind.io").Path("/headers").Request)
	Expect(out.Code).To.Equal(200)

	headers := toTyped(out.Body)
	Expect(len(headers)).To.Equal(3)
	Expect(headers.String("x-request-id")).To.Equal(id)
	Expect(headers.String("host")).To.Equal("openmymind.io")
	Expect(headers.String("accept-encoding")).To.Equal("gzip")
}

func (h *UpstreamTests) SpecificHeaders() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Host("openmymind.io").Path("/headers").Header("X-Spice", "must flow").Request)
	Expect(out.Code).To.Equal(200)

	headers := toTyped(out.Body)
	Expect(headers.String("x-spice")).To.Equal("must flow")
}

func toTyped(buffer *bytes.Buffer) typed.Typed {
	m := make(map[string]interface{})
	json.Unmarshal(buffer.Bytes(), &m)
	return typed.Typed(m)
}

func startServer() *os.Process {
	cmd := exec.Command("coffee", "server_test.coffee")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return cmd.Process
}
