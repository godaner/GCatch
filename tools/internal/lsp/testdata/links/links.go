package links

import (
	"fmt" // @link(re`".*"`,"https://godoc.org/fmt")

	"github.com/godaner/GCatch/tools/internal/lsp/foo" // @link(re`".*"`,"https://godoc.org/github.com/godaner/GCatch/tools/internal/lsp/foo")
)

var (
	_ fmt.Formatter
	_ foo.StructFoo
)
