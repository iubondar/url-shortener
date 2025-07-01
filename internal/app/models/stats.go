package models

type Stats struct {
	URLsCount  int `json:"urls"`  // количество сокращенных URL
	UsersCount int `json:"users"` // количество пользователей
}
