package queries

const (
	InsertURL string = "INSERT INTO urls (short_url, original_url) VALUES ($1, $2);"

	GetShortURL string = "SELECT short_url from urls WHERE original_url = $1;"

	GetOriginalURL string = "SELECT original_url from urls WHERE short_url = $1;"
)
