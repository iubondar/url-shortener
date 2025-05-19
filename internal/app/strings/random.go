// Пакет strings предоставляет утилиты для работы со строками.
// Включает функции для генерации случайных строк заданной длины.
package strings

import (
	"math/rand"
)

// symbols содержит набор символов, используемых для генерации случайных строк.
// Включает латинские буквы в обоих регистрах и цифры.
var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// RandString генерирует случайную строку заданной длины.
// Строка состоит из символов, определенных в переменной symbols.
//
// Параметры:
//   - n: длина генерируемой строки
//
// Возвращает:
//   - string: случайная строка заданной длины
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(b)
}
