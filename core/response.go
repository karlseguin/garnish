package core

import (
	"net/http"
)

// An interface used for a Closable response
type ByteCloser interface {
	Len() int
	Close() error
	Bytes() []byte
}

type Response interface {
	// Get the response's header
	GetHeader() http.Header

	// Get the response's body
	GetBody() []byte

	// Get the response's status
	GetStatus() int

	//set the response's status
	SetStatus(status int)

	// Close the response
	Close() error

	// Detaches the response from any underlying resourcs.
	// In cases where Close is a no-op, this should probably
	// return self. Otherwise, the response should do whatever
	// it has to so that it can be long-lived (clone itself into
	// a normal response and close itself)
	Detach() Response
}
