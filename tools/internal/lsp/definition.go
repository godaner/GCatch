// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"github.com/godaner/GCatch/tools/internal/lsp/protocol"
	"github.com/godaner/GCatch/tools/internal/lsp/source"
	"github.com/godaner/GCatch/tools/internal/span"
)

func (s *Server) definition(ctx context.Context, params *protocol.TextDocumentPositionParams) ([]protocol.Location, error) {
	uri := span.NewURI(params.TextDocument.URI)
	view := s.session.ViewOf(uri)
	f, m, err := getGoFile(ctx, view, uri)
	if err != nil {
		return nil, err
	}
	spn, err := m.PointSpan(params.Position)
	if err != nil {
		return nil, err
	}
	rng, err := spn.Range(m.Converter)
	if err != nil {
		return nil, err
	}
	ident, err := source.Identifier(ctx, view, f, rng.Start)
	if err != nil {
		return nil, err
	}
	decSpan, err := ident.Declaration.Range.Span()
	if err != nil {
		return nil, err
	}
	_, decM, err := getSourceFile(ctx, view, decSpan.URI())
	if err != nil {
		return nil, err
	}
	loc, err := decM.Location(decSpan)
	if err != nil {
		return nil, err
	}
	return []protocol.Location{loc}, nil
}

func (s *Server) typeDefinition(ctx context.Context, params *protocol.TextDocumentPositionParams) ([]protocol.Location, error) {
	uri := span.NewURI(params.TextDocument.URI)
	view := s.session.ViewOf(uri)
	f, m, err := getGoFile(ctx, view, uri)
	if err != nil {
		return nil, err
	}
	spn, err := m.PointSpan(params.Position)
	if err != nil {
		return nil, err
	}
	rng, err := spn.Range(m.Converter)
	if err != nil {
		return nil, err
	}
	ident, err := source.Identifier(ctx, view, f, rng.Start)
	if err != nil {
		return nil, err
	}
	identSpan, err := ident.Type.Range.Span()
	if err != nil {
		return nil, err
	}
	_, identM, err := getSourceFile(ctx, view, identSpan.URI())
	if err != nil {
		return nil, err
	}
	loc, err := identM.Location(identSpan)
	if err != nil {
		return nil, err
	}
	return []protocol.Location{loc}, nil
}
