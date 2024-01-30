package common

import (
	"context"
	"io"
	"io/fs"
)

// AddFunc is the type of the function called for each file or directory visited.
// The path always have '/' as path separator.
// The src is nil for directories.
// Returning error will abort the caller.
type AddFunc func(ctx context.Context, path string, info fs.FileInfo, src io.Reader) error

// ArchiveWriteToFunc is the type of the function that accepts a destination writer and
// returns the number of bytes written and an error.
type ArchiveWriteToFunc func(ctx context.Context, dest io.Writer) (int64, error)

// WriteCounter is a wrapper around io.Writer that counts the number of bytes written.
type WriteCounter struct {
	io.Writer
	N int64
}

// Write implements io.Writer.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n, err := wc.Writer.Write(p)
	if n > 0 {
		wc.N += int64(n)
	}
	return n, err
}
