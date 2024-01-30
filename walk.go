package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ozanh/arch-transfer/common"
)

func walkPath(ctx context.Context, source string, addFn common.AddFunc) error {
	source = filepath.Clean(source)

	info, err := os.Lstat(source)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}
	if info.IsDir() {
		err = walkDir(ctx, source, addFn)
	} else if info.Mode().IsRegular() {
		var f *os.File
		f, err = os.Open(source)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer f.Close()
		name := filepath.Base(source)
		err = addFn(ctx, name, info, f)
	} else {
		err = fmt.Errorf("path is not a directory or regular file: %s", source)
	}
	return err
}

func walkDir(ctx context.Context, source string, addFn common.AddFunc) error {
	parent := filepath.Dir(source)
	root := filepath.Base(source)
	if root == "" {
		root = "."
	}
	fsys := os.DirFS(parent)
	return fs.WalkDir(fsys, root, func(path string, d os.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return fmt.Errorf("failed to walk path: %s: %w", path, err)
		}
		if path == "." || !(d.IsDir() || d.Type().IsRegular()) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info: %s: %w", path, err)
		}

		var src io.Reader
		if d.Type().IsRegular() {
			f, err := fsys.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %s: %w", path, err)
			}
			defer f.Close()
			src = f
		}
		path = filepath.ToSlash(path)
		return addFn(ctx, path, info, src)
	})
}
