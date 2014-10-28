package garnish

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
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

func (h *UpstreamTests) UpstreamsARequest() {
	handler := testHandler()
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, build.Request().Path("/plain").Request)
	Expect(out.Code).To.Equal(200)
	Expect(out.HeaderMap.Get("Content-Length")).To.Equal("11")
	Expect(out.Body.String()).To.Equal("hello world")
}

func startServer() *os.Process {
	cmd := exec.Command("coffee", "server_test.coffee")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return cmd.Process
}
