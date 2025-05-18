package strings

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ExampleRandString демонстрирует использование функции RandString для генерации
// случайной строки заданной длины.
func ExampleRandString() {
	// Генерируем случайную строку длиной 8 символов
	randomStr := RandString(8)
	fmt.Println(len(randomStr))
	// Output: 8
}

// ExampleRandString_differentLengths демонстрирует генерацию строк разной длины.
func ExampleRandString_differentLengths() {
	// Генерируем строки разной длины
	str1 := RandString(5)
	str2 := RandString(10)
	fmt.Printf("Длина первой строки: %d\n", len(str1))
	fmt.Printf("Длина второй строки: %d\n", len(str2))
	// Output:
	// Длина первой строки: 5
	// Длина второй строки: 10
}

func TestSimpleRepository_SaveAndRetrieve(t *testing.T) {
	// Проверяем что вернули строку заданной длины
	a := RandString(8)
	b := RandString(8)
	c := RandString(6)

	assert.Len(t, a, 8)
	assert.Len(t, b, 8)
	assert.Len(t, c, 6)

	// Строки не одинаковые
	assert.NotEqual(t, a, b)

	// Все символы из заданного набора
	symbols := string(symbols)
	for _, r := range a {
		assert.True(t, strings.ContainsRune(string(symbols), r))
	}
}
