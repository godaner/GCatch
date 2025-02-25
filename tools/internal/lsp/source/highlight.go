// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package source

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/godaner/GCatch/tools/go/ast/astutil"
	"github.com/godaner/GCatch/tools/internal/span"
)

func Highlight(ctx context.Context, f GoFile, pos token.Pos) []span.Span {
	file := f.GetAST(ctx)
	if file == nil {
		return nil
	}
	fset := f.FileSet()
	path, _ := astutil.PathEnclosingInterval(file, pos, pos)
	if len(path) == 0 {
		return nil
	}

	id, ok := path[0].(*ast.Ident)
	if !ok {
		return nil
	}

	var result []span.Span
	if id.Obj != nil {
		ast.Inspect(path[len(path)-1], func(n ast.Node) bool {
			if n, ok := n.(*ast.Ident); ok && n.Obj == id.Obj {
				s, err := nodeSpan(n, fset)
				if err == nil {
					result = append(result, s)
				}
			}
			return true
		})
	}
	return result
}
