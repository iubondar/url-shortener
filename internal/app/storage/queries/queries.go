package queries

const (
	InsertURL string = "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3);"

	GetShortURL string = "SELECT short_url from urls WHERE original_url = $1;"

	GetByShortURL string = "SELECT user_id, short_url, original_url, is_deleted from urls WHERE short_url = $1;"

	GetUserUrls string = "SELECT user_id, short_url, original_url, is_deleted FROM urls WHERE user_id = $1;"

	DeleteUserURL string = "UPDATE urls SET is_deleted = true WHERE user_id = $1 AND short_url = $2;"
)
