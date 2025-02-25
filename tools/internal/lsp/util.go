// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"
	"fmt"

	"github.com/godaner/GCatch/tools/internal/lsp/protocol"
	"github.com/godaner/GCatch/tools/internal/lsp/source"
	"github.com/godaner/GCatch/tools/internal/span"
)

func getSourceFile(ctx context.Context, v source.View, uri span.URI) (source.File, *protocol.ColumnMapper, error) {
	f, err := v.GetFile(ctx, uri)
	if err != nil {
		return nil, nil, err
	}
	filename, err := f.URI().Filename()
	if err != nil {
		return nil, nil, err
	}
	fc := f.Content(ctx)
	if fc.Error != nil {
		return nil, nil, fc.Error
	}
	m := protocol.NewColumnMapper(f.URI(), filename, f.FileSet(), f.GetToken(ctx), fc.Data)

	return f, m, nil
}

func getGoFile(ctx context.Context, v source.View, uri span.URI) (source.GoFile, *protocol.ColumnMapper, error) {
	f, m, err := getSourceFile(ctx, v, uri)
	if err != nil {
		return nil, nil, err
	}
	gof, ok := f.(source.GoFile)
	if !ok {
		return nil, nil, fmt.Errorf("not a Go file %v", f.URI())
	}
	return gof, m, nil
}
