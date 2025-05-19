// Package queries содержит SQL-запросы для работы с таблицей urls.
// Включает в себя запросы для:
// - Добавления новых URL
// - Получения короткого URL по оригинальному
// - Получения информации по короткому URL
// - Получения всех URL пользователя
// - Мягкого удаления URL пользователя
package queries

// SQL-запросы для работы с таблицей urls.
const (
	// InsertURL добавляет новую запись в таблицу urls.
	// Параметры:
	// $1 - короткий URL
	// $2 - оригинальный URL
	// $3 - ID пользователя
	InsertURL string = "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3);"

	// GetShortURL возвращает короткий URL по оригинальному URL.
	// Параметры:
	// $1 - оригинальный URL
	GetShortURL string = "SELECT short_url from urls WHERE original_url = $1;"

	// GetByShortURL возвращает полную информацию о URL по его короткой версии.
	// Параметры:
	// $1 - короткий URL
	GetByShortURL string = "SELECT user_id, short_url, original_url, is_deleted from urls WHERE short_url = $1;"

	// GetUserUrls возвращает все URL, принадлежащие пользователю.
	// Параметры:
	// $1 - ID пользователя
	GetUserUrls string = "SELECT user_id, short_url, original_url, is_deleted FROM urls WHERE user_id = $1;"

	// DeleteUserURL выполняет мягкое удаление URL пользователя.
	// Параметры:
	// $1 - ID пользователя
	// $2 - короткий URL
	DeleteUserURL string = "UPDATE urls SET is_deleted = true WHERE user_id = $1 AND short_url = $2;"
)
