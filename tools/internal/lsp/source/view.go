// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package source

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/godaner/GCatch/tools/go/analysis"
	"github.com/godaner/GCatch/tools/go/packages"
	"github.com/godaner/GCatch/tools/internal/lsp/diff"
	"github.com/godaner/GCatch/tools/internal/lsp/xlog"
	"github.com/godaner/GCatch/tools/internal/span"
)

// FileContents is returned from FileSystem implementation to represent the
// contents of a file.
type FileContent struct {
	URI   span.URI
	Data  []byte
	Error error
	Hash  string
}

// FileSystem is the interface to something that provides file contents.
type FileSystem interface {
	// ReadFile reads the contents of a file and returns it.
	ReadFile(uri span.URI) *FileContent
}

// Cache abstracts the core logic of dealing with the environment from the
// higher level logic that processes the information to produce results.
// The cache provides access to files and their contents, so the source
// package does not directly access the file system.
// A single cache is intended to be process wide, and is the primary point of
// sharing between all consumers.
// A cache may have many active sessions at any given time.
type Cache interface {
	// A FileSystem that reads file contents from external storage.
	FileSystem

	// NewSession creates a new Session manager and returns it.
	NewSession(log xlog.Logger) Session

	// FileSet returns the shared fileset used by all files in the system.
	FileSet() *token.FileSet
}

// Session represents a single connection from a client.
// This is the level at which things like open files are maintained on behalf
// of the client.
// A session may have many active views at any given time.
type Session interface {
	// NewView creates a new View and returns it.
	NewView(name string, folder span.URI) View

	// Cache returns the cache that created this session.
	Cache() Cache

	// Returns the logger in use for this session.
	Logger() xlog.Logger

	// View returns a view with a mathing name, if the session has one.
	View(name string) View

	// ViewOf returns a view corresponding to the given URI.
	ViewOf(uri span.URI) View

	// Views returns the set of active views built by this session.
	Views() []View

	// Shutdown the session and all views it has created.
	Shutdown(ctx context.Context)

	// A FileSystem prefers the contents from overlays, and falls back to the
	// content from the underlying cache if no overlay is present.
	FileSystem

	// DidOpen is invoked each time a file is opened in the editor.
	DidOpen(uri span.URI)

	// DidSave is invoked each time an open file is saved in the editor.
	DidSave(uri span.URI)

	// DidClose is invoked each time an open file is closed in the editor.
	DidClose(uri span.URI)

	// IsOpen can be called to check if the editor has a file currently open.
	IsOpen(uri span.URI) bool

	// Called to set the effective contents of a file from this session.
	SetOverlay(uri span.URI, data []byte)
}

// View represents a single workspace.
// This is the level at which we maintain configuration like working directory
// and build tags.
type View interface {
	// Session returns the session that created this view.
	Session() Session

	// Name returns the name this view was constructed with.
	Name() string

	// Folder returns the root folder for this view.
	Folder() span.URI

	// BuiltinPackage returns the ast for the special "builtin" package.
	BuiltinPackage() *ast.Package

	// GetFile returns the file object for a given uri.
	GetFile(ctx context.Context, uri span.URI) (File, error)

	// Called to set the effective contents of a file from this view.
	SetContent(ctx context.Context, uri span.URI, content []byte) error

	// BackgroundContext returns a context used for all background processing
	// on behalf of this view.
	BackgroundContext() context.Context

	// Env returns the current set of environment overrides on this view.
	Env() []string

	// SetEnv is used to adjust the environment applied to the view.
	SetEnv([]string)

	// SetBuildFlags is used to adjust the build flags applied to the view.
	SetBuildFlags([]string)

	// Shutdown closes this view, and detaches it from it's session.
	Shutdown(ctx context.Context)

	// Ignore returns true if this file should be ignored by this view.
	Ignore(span.URI) bool
}

// File represents a source file of any type.
type File interface {
	URI() span.URI
	View() View
	Content(ctx context.Context) *FileContent
	FileSet() *token.FileSet
	GetToken(ctx context.Context) *token.File
}

// GoFile represents a Go source file that has been type-checked.
type GoFile interface {
	File

	// GetTrimmedAST returns an AST that may or may not contain function bodies.
	// It should be used in scenarios where function bodies are not necessary.
	GetTrimmedAST(ctx context.Context) *ast.File

	// GetAST returns the full AST for the file.
	GetAST(ctx context.Context) *ast.File

	// GetPackage returns the package that this file belongs to.
	GetPackage(ctx context.Context) Package

	// GetActiveReverseDeps returns the active files belonging to the reverse
	// dependencies of this file's package.
	GetActiveReverseDeps(ctx context.Context) []GoFile
}

type ModFile interface {
	File
}

type SumFile interface {
	File
}

// Package represents a Go package that has been type-checked. It maintains
// only the relevant fields of a *go/packages.Package.
type Package interface {
	PkgPath() string
	GetFilenames() []string
	GetSyntax() []*ast.File
	GetErrors() []packages.Error
	GetTypes() *types.Package
	GetTypesInfo() *types.Info
	GetTypesSizes() types.Sizes
	IsIllTyped() bool
	GetActionGraph(ctx context.Context, a *analysis.Analyzer) (*Action, error)
	GetImport(pkgPath string) Package
}

// TextEdit represents a change to a section of a document.
// The text within the specified span should be replaced by the supplied new text.
type TextEdit struct {
	Span    span.Span
	NewText string
}

// DiffToEdits converts from a sequence of diff operations to a sequence of
// source.TextEdit
func DiffToEdits(uri span.URI, ops []*diff.Op) []TextEdit {
	edits := make([]TextEdit, 0, len(ops))
	for _, op := range ops {
		s := span.New(uri, span.NewPoint(op.I1+1, 1, 0), span.NewPoint(op.I2+1, 1, 0))
		switch op.Kind {
		case diff.Delete:
			// Delete: unformatted[i1:i2] is deleted.
			edits = append(edits, TextEdit{Span: s})
		case diff.Insert:
			// Insert: formatted[j1:j2] is inserted at unformatted[i1:i1].
			if content := strings.Join(op.Content, ""); content != "" {
				edits = append(edits, TextEdit{Span: s, NewText: content})
			}
		}
	}
	return edits
}

func EditsToDiff(edits []TextEdit) []*diff.Op {
	iToJ := 0
	ops := make([]*diff.Op, len(edits))
	for i, edit := range edits {
		i1 := edit.Span.Start().Line() - 1
		i2 := edit.Span.End().Line() - 1
		kind := diff.Insert
		if edit.NewText == "" {
			kind = diff.Delete
		}
		ops[i] = &diff.Op{
			Kind:    kind,
			Content: diff.SplitLines(edit.NewText),
			I1:      i1,
			I2:      i2,
			J1:      i1 + iToJ,
		}
		if kind == diff.Insert {
			iToJ += len(ops[i].Content)
		} else {
			iToJ -= i2 - i1
		}
	}
	return ops
}
