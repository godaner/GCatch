package bmoc

import (
	"fmt"
	"github.com/godaner/GCatch/analysis/pointer"
	"github.com/godaner/GCatch/config"
	"github.com/godaner/GCatch/instinfo"
	"github.com/godaner/GCatch/syncgraph"
	"strconv"
)

func Detect() {
	stPtrResult, vecStOpValue := pointer.AnalyzeAllSyncOp()
	if stPtrResult == nil || vecStOpValue == nil {
		return
	}
	vecChannel := pointer.WithdrawAllChan(stPtrResult, vecStOpValue)
	vecLocker := pointer.WithdrawAllTraditionals(stPtrResult, vecStOpValue)

	mapDependency := syncgraph.GenDMap(vecChannel, vecLocker)

	for _, ch := range vecChannel {
		//p := config.Prog.Fset.Position(ch.MakeInst.Pos())
		//_ = p
		if OKToCheck(ch) == true {
			CheckCh(ch, vecChannel, vecLocker, mapDependency)
		}
	}

}

var countCh int
var countUnbufferBug int
var countBufferBug int

func OKToCheck(ch *instinfo.Channel) (boolCheck bool) {
	boolCheck = false

	if ch.MakeInst == nil {
		return
	}
	pkg := ch.MakeInst.Parent().Pkg
	if pkg == nil {
		return
	}
	pkgOfPkg := pkg.Pkg
	if pkgOfPkg == nil {
		return
	}
	if config.IsPathIncluded(pkgOfPkg.Path()) == false {
		return
	}

	p := config.Prog.Fset.Position(ch.MakeInst.Pos())
	strChHash := ch.MakeInst.Parent().String() + ch.MakeInst.String() + ch.MakeInst.Name() + strconv.Itoa(p.Line)
	if _, checked := config.MapHashOfCheckedCh[strChHash]; checked {
		return
	}

	boolCheck = true
	config.MapHashOfCheckedCh[strChHash] = struct{}{}
	countCh++
	return
}

func CheckCh(ch *instinfo.Channel, vecChannel []*instinfo.Channel, vecLocker []*instinfo.Locker, mapDependency map[interface{}]*syncgraph.DPrim) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	syncGraph, err := syncgraph.BuildGraph(ch, vecChannel, vecLocker, mapDependency)
	if err != nil { // Met some error
		if config.Print_Debug_Info {
			fmt.Println("-----count_ch:", countCh)
		}
		return
	}

	syncGraph.ComputeFnOnOpPath()
	syncGraph.OptimizeBB_V1()

	syncGraph.SetEnumCfg(1, false, true)

	syncGraph.EnumerateAllPathCombinations()

	if ch.Buffer == instinfo.DynamicSize {
		// If this is a buffered channel with dynamic size and no critical section is found, skip this channel
	} else {
		found_GL := syncGraph.CheckWithZ3()
		if found_GL {
			if ch.Buffer == 0 {
				countUnbufferBug++
			} else {
				countBufferBug++
			}
		}
		if config.Print_Debug_Info {
			fmt.Println("-----count_unbuffer_bug:", countUnbufferBug,"---buffer_bug:", countBufferBug)
		}
	}
	if config.Print_Debug_Info {
		fmt.Println("-----count_ch:", countCh)
	}
	return
}