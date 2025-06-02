// Package main предоставляет инструмент статического анализа кода, который объединяет
// несколько анализаторов для проверки Go-кода на соответствие различным правилам и лучшим практикам.
//
// Механизм работы multichecker:
// Multichecker объединяет несколько анализаторов в один инструмент. Каждый анализатор
// представляет собой отдельную проверку кода, которая может выявлять определенные
// проблемы или предлагать улучшения. При запуске инструмента все анализаторы
// применяются последовательно к указанным файлам или пакетам.
//
// Используемые анализаторы:
//
// Стандартные анализаторы (golang.org/x/tools/go/analysis/passes):
// - asmdecl: проверяет корректность объявлений ассемблерных файлов
// - assign: выявляет недопустимые присваивания
// - atomic: проверяет правильность использования атомарных операций
// - bools: находит подозрительные операции с булевыми значениями
// - buildtag: проверяет корректность тегов сборки
// - cgocall: выявляет проблемы в вызовах CGO
// - composite: проверяет корректность составных литералов
// - copylock: находит копирование мьютексов
// - errorsas: проверяет правильность использования errors.As
// - httpresponse: выявляет проблемы с обработкой HTTP-ответов
// - loopclosure: находит проблемы с замыканиями в циклах
// - lostcancel: выявляет утечки контекстов
// - nilfunc: проверяет сравнения с nil
// - printf: проверяет форматирование строк
// - shift: выявляет проблемы с побитовыми сдвигами
// - sigchanyzer: проверяет корректность использования каналов
// - stdmethods: проверяет соответствие стандартным интерфейсам
// - stringintconv: выявляет проблемы преобразования строк в числа
// - structtag: проверяет корректность тегов структур
// - testinggoroutine: находит проблемы с горутинами в тестах
// - tests: проверяет корректность тестов
// - unmarshal: выявляет проблемы с анмаршалингом
// - unreachable: находит недостижимый код
// - unsafeptr: проверяет корректность использования unsafe.Pointer
// - unusedresult: выявляет неиспользуемые результаты функций
//
// Дополнительные анализаторы:
// - staticcheck: набор правил для улучшения качества кода
//   - SA*: группа проверок, выявляющих потенциальные ошибки и проблемы с производительностью, включая неправильное использование стандартной библиотеки, проблемы с конкурентностью и неоптимальные паттерны кода
//   - S1025: предупреждает о ненужном использовании fmt.Sprintf("%s", x)
//   - S1017: рекомендует использовать strings.TrimPrefix вместо ручной обрезки
//
// - simple: набор правил для упрощения кода
// - errcheck: проверяет обработку ошибок
// - os.Exit: проверяет использование os.Exit в main
//
// Запуск:
// go run cmd/staticlint/main.go ./...
package main

import (
	"strings"

	"github.com/iubondar/url-shortener/internal/analyser"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
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

	// Добавляем наш анализатор os.Exit
	mychecks = append(mychecks, analyser.OsExitAnalyzer)

	// Добавляем все стандартные анализаторы
	mychecks = append(mychecks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	multichecker.Main(
		mychecks...,
	)
}
