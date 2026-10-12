package main

import (
	"github.com/hrdtr/readable-gcl/pkg/namingclarity"
	"github.com/hrdtr/readable-gcl/pkg/testconvention"
	"golang.org/x/tools/go/analysis/multichecker"
)

var multicheckerMain = multichecker.Main

func main() {
	multicheckerMain(
		namingclarity.Analyzer,
		testconvention.Analyzer,
	)
}
