package cache

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/godaner/GCatch/tools/internal/lsp/source"
	"github.com/godaner/GCatch/tools/internal/span"
)

// goFile holds all of the information we know about a Go file.
type goFile struct {
	fileBase

	ast *astFile

	pkg     *pkg
	meta    *metadata
	imports []*ast.ImportSpec
}

type astFile struct {
	file      *ast.File
	isTrimmed bool
}

func (f *goFile) GetToken(ctx context.Context) *token.File {
	f.view.mu.Lock()
	defer f.view.mu.Unlock()

	if f.isDirty() {
		if _, err := f.view.loadParseTypecheck(ctx, f); err != nil {
			f.View().Session().Logger().Errorf(ctx, "unable to check package for %s: %v", f.URI(), err)
			return nil
		}
	}
	return f.token
}

func (f *goFile) GetTrimmedAST(ctx context.Context) *ast.File {
	f.view.mu.Lock()
	defer f.view.mu.Unlock()

	if f.isDirty() {
		if _, err := f.view.loadParseTypecheck(ctx, f); err != nil {
			f.View().Session().Logger().Errorf(ctx, "unable to check package for %s: %v", f.URI(), err)
			return nil
		}
	}
	return f.ast.file
}

func (f *goFile) GetAST(ctx context.Context) *ast.File {
	f.view.mu.Lock()
	defer f.view.mu.Unlock()

	if f.isDirty() || f.astIsTrimmed() {
		if _, err := f.view.loadParseTypecheck(ctx, f); err != nil {
			f.View().Session().Logger().Errorf(ctx, "unable to check package for %s: %v", f.URI(), err)
			return nil
		}
	}
	return f.ast.file
}

func (f *goFile) GetPackage(ctx context.Context) source.Package {
	f.view.mu.Lock()
	defer f.view.mu.Unlock()

	if f.isDirty() || f.astIsTrimmed() {
		if errs, err := f.view.loadParseTypecheck(ctx, f); err != nil {
			f.View().Session().Logger().Errorf(ctx, "unable to check package for %s: %v", f.URI(), err)

			// Create diagnostics for errors if we are able to.
			if len(errs) > 0 {
				return &pkg{errors: errs}
			}
			return nil
		}
	}
	return f.pkg
}

// isDirty is true if the file needs to be type-checked.
// It assumes that the file's view's mutex is held by the caller.
func (f *goFile) isDirty() bool {
	return f.meta == nil || f.imports == nil || f.token == nil || f.ast == nil || f.pkg == nil || len(f.view.contentChanges) > 0
}

func (f *goFile) astIsTrimmed() bool {
	return f.ast != nil && f.ast.isTrimmed
}

func (f *goFile) GetActiveReverseDeps(ctx context.Context) []source.GoFile {
	pkg := f.GetPackage(ctx)
	if pkg == nil {
		return nil
	}

	f.view.mu.Lock()
	defer f.view.mu.Unlock()

	f.view.mcache.mu.Lock()
	defer f.view.mcache.mu.Unlock()

	seen := make(map[string]struct{}) // visited packages
	results := make(map[*goFile]struct{})
	f.view.reverseDeps(ctx, seen, results, pkg.PkgPath())

	var files []source.GoFile
	for rd := range results {
		if rd == nil {
			continue
		}
		// Don't return any of the active files in this package.
		if rd.pkg != nil && rd.pkg == pkg {
			continue
		}
		files = append(files, rd)
	}
	return files
}

func (v *view) reverseDeps(ctx context.Context, seen map[string]struct{}, results map[*goFile]struct{}, pkgPath string) {
	if _, ok := seen[pkgPath]; ok {
		return
	}
	seen[pkgPath] = struct{}{}
	m, ok := v.mcache.packages[pkgPath]
	if !ok {
		return
	}
	for _, filename := range m.files {
		uri := span.FileURI(filename)
		if f, err := v.getFile(uri); err == nil && v.session.IsOpen(uri) {
			results[f.(*goFile)] = struct{}{}
		}
	}
	for parentPkgPath := range m.parents {
		v.reverseDeps(ctx, seen, results, parentPkgPath)
	}
}
