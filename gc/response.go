package gc

import (
	"net/http"
)

var (
	NotFoundResponse = Empty(404)
)

type ByteCloser interface {
	Bytes() []byte
	Close() error
}

type Response interface {
	Body() []byte
	Status() int
	Header() http.Header
	Close()
}

func Empty(status int) Response {
	return &NormalResponse{
		status: status,
	}
}

func Respond(status int, body interface{}) Response {
	return RespondH(status, nil, body)
}

func RespondH(status int, header http.Header, body interface{}) Response {
	switch b := body.(type) {
	case string:
		return &NormalResponse{[]byte(b), status, header}
	case []byte:
		return &NormalResponse{b, status, header}
	case ByteCloser:
		return &CloseableResponse{b, status, header}
	default:
		return Fatal("invalid body type")
	}
}

type NormalResponse struct {
	body   []byte
	status int
	header http.Header
}

func (r *NormalResponse) Body() []byte {
	return r.body
}

func (r *NormalResponse) Status() int {
	return r.status
}

func (r *NormalResponse) Header() http.Header {
	return r.header
}

func (r *NormalResponse) Close() {

}

type FatalResponse struct {
	Response
	Err string
}

func Fatal(message string) Response {
	return &FatalResponse{
		Err:      message,
		Response: Empty(500),
	}
}

func FatalErr(err error) Response {
	return &FatalResponse{
		Err:      err.Error(),
		Response: Empty(500),
	}
}

type CloseableResponse struct {
	body   ByteCloser
	status int
	header http.Header
}

func (r *CloseableResponse) Body() []byte {
	return r.body.Bytes()
}

func (r *CloseableResponse) Status() int {
	return r.status
}

func (r *CloseableResponse) Header() http.Header {
	return r.header
}

func (r *CloseableResponse) Close() {
	r.body.Close()
}
