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

func (s *Server) documentHighlight(ctx context.Context, params *protocol.TextDocumentPositionParams) ([]protocol.DocumentHighlight, error) {
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
	spans := source.Highlight(ctx, f, rng.Start)
	return toProtocolHighlight(m, spans), nil
}

func toProtocolHighlight(m *protocol.ColumnMapper, spans []span.Span) []protocol.DocumentHighlight {
	result := make([]protocol.DocumentHighlight, 0, len(spans))
	kind := protocol.Text
	for _, span := range spans {
		r, err := m.Range(span)
		if err != nil {
			continue
		}
		h := protocol.DocumentHighlight{Kind: &kind, Range: r}
		result = append(result, h)
	}
	return result
}
