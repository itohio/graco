package processor

import (
	"io"
	"sync"
)

// Blob interface represents a type that is an intermediary between actual type and
// bytes array. For example it can be a result of a JSON or Protobuf marshaller.
type Blob[T any] interface {
	io.Closer
	Val() T
	Data() []byte
}

type bytesBlob[T any] struct {
	bytes []byte
	val   T
}

func (b bytesBlob[T]) Val() T       { return b.val }
func (b bytesBlob[T]) Data() []byte { return b.bytes }
func (b bytesBlob[T]) Close() error { return nil }

func NewBlob[T any](val T, bytes []byte) bytesBlob[T] {
	return bytesBlob[T]{
		bytes: bytes,
		val:   val,
	}
}

type poolBlob[T any] struct {
	bytesBlob[T]
	p *sync.Pool
}

func (b poolBlob[T]) Val() T       { return b.val }
func (b poolBlob[T]) Data() []byte { return b.bytes }
func (b *poolBlob[T]) Close() error {
	bytes := b.bytes
	pool := b.p
	b.bytes = nil
	b.p = nil
	var zero T
	b.val = zero
	pool.Put(bytes)
	return nil
}

func NewBlobWithPool[T any](p *sync.Pool, val T, bytes []byte) poolBlob[T] {
	return poolBlob[T]{
		bytesBlob: bytesBlob[T]{
			bytes: bytes,
			val:   val,
		},
		p: p,
	}
}
