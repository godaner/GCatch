// The findcall command runs the findcall analyzer.
package main

import (
	"github.com/godaner/GCatch/tools/go/analysis/passes/findcall"
	"github.com/godaner/GCatch/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(findcall.Analyzer) }
