// Package analyser предоставляет инструменты для статического анализа Go кода.
package analyser

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer - анализатор, который запрещает прямое использование os.Exit в функции main пакета main.
// Анализатор проверяет все файлы в проекте и выдает предупреждение, если обнаружит вызов os.Exit
// в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "os_exit",
	Doc:  "prohibit os.Exit direct use in main package main function",
	Run:  run,
}

// checkOsExitCalls проверяет наличие вызовов os.Exit в указанной функции.
// Функция рекурсивно обходит AST функции и ищет вызовы os.Exit.
// При обнаружении вызова выдает предупреждение через pass.Reportf.
func checkOsExitCalls(pass *analysis.Pass, fd *ast.FuncDecl) {
	ast.Inspect(fd, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "os.Exit directly called in main function")
			}
		}

		return true
	})
}

// run является основной функцией анализатора.
// Она обходит все файлы в проекте, пропуская файлы из кэша сборки,
// и проверяет функцию main на наличие вызовов os.Exit.
func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		// Skip files from cache
		if strings.Contains(filepath.Dir(pass.Fset.File(f.Pos()).Name()), "go-build") {
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			// Check if we're in a main function
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Name.Name != "main" {
				return true
			}

			checkOsExitCalls(pass, fd)
			return true
		})
	}

	return nil, nil
}
