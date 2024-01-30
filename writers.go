package main

import (
	"context"
	"io"

	"github.com/ozanh/arch-transfer/common"
	"github.com/ozanh/arch-transfer/ziparch"
)

func zipArchiveWriteTo(source string) common.ArchiveWriteToFunc {
	return func(ctx context.Context, dest io.Writer) (int64, error) {
		counter := &common.WriteCounter{Writer: dest}
		za := ziparch.NewZipArchive(counter)
		err := walkPath(ctx, source, za.Add)
		if err != nil {
			return counter.N, err
		}
		err = za.Close()
		return counter.N, err
	}
}
