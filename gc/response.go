package gc

import (
	"github.com/karlseguin/params"
)

type Response interface {
	Body() []byte
	Status() int
	Header() params.Params
	Close()
}

func Empty(status int) Response {
	return &NormalResponse{
		status: status,
		header: params.Empty,
	}
}

type NormalResponse struct {
	body   []byte
	status int
	header params.Params
}

func (r *NormalResponse) Body() []byte {
	return r.body
}

func (r *NormalResponse) Status() int {
	return r.status
}

func (r *NormalResponse) Header() params.Params {
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
