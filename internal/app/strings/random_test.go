package strings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
