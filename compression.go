package micup

import (
	"bytes"
	"io"
)

// PayloadDecompressor interface for decompressing gateway packets.
type PayloadDecompressor interface {
	Decompress(data []byte) ([]byte, error)
}

// PlainDecompressor is a no-op decompressor for standard JSON encoding.
type PlainDecompressor struct{}

func (p *PlainDecompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

// NewPlainDecompressor returns a pass-through decompressor.
func NewPlainDecompressor() PayloadDecompressor {
	return &PlainDecompressor{}
}

// BufferDecompressor wraps an in-memory buffer reader.
type BufferDecompressor struct {
	buf bytes.Buffer
}

// Reset clears the buffer.
func (b *BufferDecompressor) Reset() {
	b.buf.Reset()
}

// Write appends chunk bytes.
func (b *BufferDecompressor) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

// Reader returns reading access.
func (b *BufferDecompressor) Reader() io.Reader {
	return &b.buf
}
