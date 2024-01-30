package ziparch

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
)

// ZipArchive is a struct to create zip archive for given files and directories
// via Add(). Close() must be called after all files are added successfully.
type ZipArchive struct {
	zw *zip.Writer
}

// NewZipArchive returns a new ZipArchive.
func NewZipArchive(dest io.Writer) *ZipArchive {
	return &ZipArchive{
		zw: zip.NewWriter(dest),
	}
}

// Add creates a zip header from info and adds the file to the zip archive.
// If given path is a directory, it will be added as a directory and src will be
// ignored.
func (za *ZipArchive) Add(ctx context.Context, path string, info fs.FileInfo, src io.Reader) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create header from info: %w", err)
	}
	header.Name = path

	w, err := za.zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	_, err = io.Copy(w, src)
	if err != nil {
		return fmt.Errorf("failed to copy file to zip archive: %w", err)
	}
	return nil
}

// Close closes the zip archive to flush the contents and finalize the zip file.
func (za *ZipArchive) Close() error {
	return za.zw.Close()
}
