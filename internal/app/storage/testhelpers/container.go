// Пакет testhelpers предоставляет утилиты для тестирования, включая управление контейнерами
// для интеграционных тестов. Использует testcontainers-go для управления тестовыми зависимостями,
// такими как базы данных PostgreSQL.
package testhelpers

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer оборачивает PostgreSQL контейнер из testcontainers-go
// и предоставляет дополнительную функциональность для целей тестирования.
type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

// CreatePostgresContainer создает и запускает новый контейнер PostgreSQL для тестирования.
// Использует образ PostgreSQL 15.3 Alpine и настраивает его с учетными данными по умолчанию для тестов.
//
// Контейнер настроен со следующими параметрами:
// - Имя базы данных: test-db
// - Имя пользователя: postgres
// - Пароль: postgres
// - Режим SSL: отключен
//
// Функция ожидает готовности базы данных к приему соединений перед возвратом.
//
// Параметры:
//   - ctx: Контекст для управления жизненным циклом контейнера
//
// Возвращает:
//   - *PostgresContainer: Экземпляр контейнера с деталями подключения
//   - error: Любая ошибка, возникшая при создании контейнера
func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}
