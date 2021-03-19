package nodisk

import (
	"github.com/godaner/GCatch/tools/internal/lsp/foo"
)

func _() {
	foo.Foo() //@complete("F", Foo, IntFoo, StructFoo)
}
