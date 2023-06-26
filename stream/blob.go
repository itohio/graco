package stream

import "io"

// Blob interface represents a type that is an intermediary between actual type and
// bytes array. For example it can be a result of a JSON or Protobuf marshaller.
type Blob interface {
	io.Closer
	Val() any
	Data() []byte
}
