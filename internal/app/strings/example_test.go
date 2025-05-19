package strings

import (
	"fmt"
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
