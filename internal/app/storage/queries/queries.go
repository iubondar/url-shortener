package queries

const (
	CreateURLsTable string = "CREATE TABLE IF NOT EXISTS urls (" +
		"id SERIAL PRIMARY KEY," +
		"short_url VARCHAR(10) UNIQUE NOT NULL," +
		"original_url VARCHAR(2048) UNIQUE NOT NULL" +
		");"

	CreateShortURLIndex string = "CREATE UNIQUE INDEX IF NOT EXISTS short_url_index ON urls (short_url)"

	CreateOriginalURLIndex string = "CREATE UNIQUE INDEX IF NOT EXISTS original_url_index ON urls (original_url)"

	InsertURL string = "INSERT INTO urls (short_url, original_url) VALUES ($1, $2);"

	GetShortURL string = "SELECT short_url from urls WHERE original_url = $1;"

	GetOriginalURL string = "SELECT original_url from urls WHERE short_url = $1;"
)
