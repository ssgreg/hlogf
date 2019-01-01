package main

import (
	"sync"

	"github.com/ssgreg/logf"
)

// NewPool returns a new Pool.
func NewPool() Pool {
	return Pool{p: &sync.Pool{
		New: func() interface{} {
			return logf.NewBufferWithCapacity(1024)
		},
	}}
}

// A Pool is a type-safe wrapper around a sync.Pool.
type Pool struct {
	p *sync.Pool
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p Pool) Get() *logf.Buffer {
	buf := p.p.Get().(*logf.Buffer)
	buf.Reset()

	return buf
}

// Put puts a Buffer back to the pool.
func (p Pool) Put(buf *logf.Buffer) {
	p.p.Put(buf)
}
