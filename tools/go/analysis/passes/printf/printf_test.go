package printf_test

import (
	"testing"

	"github.com/godaner/GCatch/tools/go/analysis/analysistest"
	"github.com/godaner/GCatch/tools/go/analysis/passes/printf"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	printf.Analyzer.Flags.Set("funcs", "Warn,Warnf")
	analysistest.Run(t, testdata, printf.Analyzer, "a", "b", "nofmt")
}
