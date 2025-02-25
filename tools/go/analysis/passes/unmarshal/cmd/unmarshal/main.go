// The unmarshal command runs the unmarshal analyzer.
package main

import (
	"github.com/godaner/GCatch/tools/go/analysis/passes/unmarshal"
	"github.com/godaner/GCatch/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(unmarshal.Analyzer) }
