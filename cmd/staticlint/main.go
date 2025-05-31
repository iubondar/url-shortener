package main

import (
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// определяем map подключаемых дополнительных правил
	checks := map[string]bool{
		"S1025": true, // Don't use fmt.Sprintf("%s", x) unnecessarily
		"S1017": true, // Replace manual trimming with strings.TrimPrefix
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки:
		// - дополнительные правила
		// - все правила из SA пакета
		if checks[v.Analyzer.Name] || strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Добавляем анализатор gosimple
	for _, a := range simple.Analyzers {
		mychecks = append(mychecks, a.Analyzer)
	}

	// Добавляем анализатор errcheck
	mychecks = append(mychecks, errcheck.Analyzer)

	multichecker.Main(
		mychecks...,
	)
}
