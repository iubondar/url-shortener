package queries

const (
	InsertURL string = "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3);"

	GetShortURL string = "SELECT short_url from urls WHERE original_url = $1;"

	GetOriginalURL string = "SELECT original_url from urls WHERE short_url = $1;"

	GetUserUrls string = "SELECT short_url, original_url FROM urls WHERE user_id = $1;"
)
