package b

import (
	"fmt"

	"github.com/godaner/GCatch/tools/internal/lsp/godef/a"
)

func useThings() {
	t := a.Thing{}      //@mark(bStructType, "ing")
	fmt.Print(t.Member) //@mark(bMember, "ember")
	fmt.Print(a.Other)  //@mark(bVar, "ther")
	a.Things()          //@mark(bFunc, "ings")
}

/*@
godef(bStructType, Thing)
godef(bMember, Member)
godef(bVar, Other)
godef(bFunc, Things)
*/
