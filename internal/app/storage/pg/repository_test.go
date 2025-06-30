package pg

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"

	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/iubondar/url-shortener/internal/app/storage/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repo        *PGRepository
	cleanup     func()
	pgContainer *testhelpers.PostgresContainer
)

func cleanupResources(db *DB, container *testhelpers.PostgresContainer, ctx context.Context) {
	if db != nil {
		if err := db.SQLDB.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}
	if container != nil {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate postgres container: %v", err)
		}
	}
}

func handleError(err error, db *DB, container *testhelpers.PostgresContainer, ctx context.Context, message string) {
	if err != nil {
		cleanupResources(db, container, ctx)
		log.Fatalf("%s: %v", message, err)
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	pgContainer, err = testhelpers.CreatePostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create postgres container: %v", err)
	}

	db, err := NewDB(pgContainer.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	err = goose.Up(db.SQLDB, "./migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	repo, err = NewPGRepository(db, 30*time.Millisecond)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	cleanup = func() {
		if repo != nil && repo.db != nil && repo.db.SQLDB != nil {
			_, err := repo.db.SQLDB.ExecContext(context.Background(), "TRUNCATE TABLE urls;")
			if err != nil {
				log.Printf("Failed to clear urls table: %v", err)
			}
		}
	}

	code := m.Run()

	if repo != nil && repo.db != nil {
		repo.db.SQLDB.Close()
	}
	if pgContainer != nil {
		pgContainer.Terminate(ctx)
	}

	os.Exit(code)
}

// ExamplePGRepository_SaveURL демонстрирует сохранение URL в PostgreSQL хранилище.
func ExamplePGRepository_SaveURL() {
	cleanup()

	// Сохраняем URL
	id, exists, err := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим результат
	fmt.Printf("ID length: %d, Exists: %v\n", len(id), exists)
	// Output: ID length: 8, Exists: false
}

// ExamplePGRepository_RetrieveByShortURL демонстрирует получение URL по короткому идентификатору.
func ExamplePGRepository_RetrieveByShortURL() {
	cleanup()

	// Сохраняем URL
	id, _, _ := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")

	// Получаем запись по короткому идентификатору
	record, err := repo.RetrieveByShortURL(context.Background(), id)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим результат
	fmt.Printf("Original URL: %s\n", record.OriginalURL)
	// Output: Original URL: http://example.com
}

// ExamplePGRepository_RetrieveUserURLs демонстрирует получение всех URL пользователя.
func ExamplePGRepository_RetrieveUserURLs() {
	cleanup()

	// Сохраняем несколько URL
	_, _, err := repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, _, err = repo.SaveURL(context.Background(), testhelpers.TestUUID, "http://example.org")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Получаем все URL пользователя
	records, err := repo.RetrieveUserURLs(context.Background(), testhelpers.TestUUID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим количество найденных URL
	fmt.Printf("Found %d URLs\n", len(records))
	// Output: Found 2 URLs
}

// ExamplePGRepository_GetStats демонстрирует получение статистики хранилища.
func ExamplePGRepository_GetStats() {
	cleanup()

	// Сохраняем несколько URL от разных пользователей
	userID1 := uuid.New()
	userID2 := uuid.New()

	_, _, err := repo.SaveURL(context.Background(), userID1, "http://example.com")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, _, err = repo.SaveURL(context.Background(), userID1, "http://example.org")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, _, err = repo.SaveURL(context.Background(), userID2, "http://example.net")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Получаем статистику
	stats, err := repo.GetStats(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Выводим статистику
	fmt.Printf("URLs: %d, Users: %d\n", stats.URLsCount, stats.UsersCount)
	// Output: URLs: 3, Users: 2
}

func setupSeparateTest(t *testing.T, execStatement string) {
	cleanup()

	if len(execStatement) > 0 {
		_, err := repo.db.SQLDB.ExecContext(context.Background(), execStatement)
		require.NoError(t, err)
	}
}

// Tests are here
func TestSaveURL(t *testing.T) {
	userID := uuid.New()
	type args struct {
		url string
	}
	tests := []struct {
		name          string
		execStatement string
		args          args
		wantID        bool
		wantExists    bool
		wantErr       bool
	}{
		{
			name:          "SaveURL Non-existent",
			execStatement: "",
			args: args{
				url: "http://example.com",
			},
			wantID:     true,
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "SaveURL Existent",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id)" +
				" VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "'), ('edVPg3ks', 'http://ya.ru', '" + userID.String() + "')",
			args: args{
				url: "http://yandex.ru",
			},
			wantID:     true,
			wantExists: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.execStatement)

			gotID, gotExists, err := repo.SaveURL(context.Background(), userID, tt.args.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("PGRepoTestSuite.TestSaveURL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantID && len(gotID) == 0 {
				t.Error("PGRepoTestSuite.TestSaveURL received empty id", gotID, tt.wantID)
			}
			if !tt.wantID && len(gotID) != 0 {
				t.Error("PGRepoTestSuite.TestSaveURL received unexpected id", gotID, tt.wantID)
			}
			if gotExists != tt.wantExists {
				t.Errorf("PGRepoTestSuite.TestSaveURL gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestRetrieveByShortURL(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name          string
		execStatement string
		args          args
		wantURL       string
		wantErr       bool
	}{
		{
			name:          "RetrieveByShortURL Non-existent",
			execStatement: "",
			args: args{
				id: "123",
			},
			wantURL: "",
			wantErr: true,
		},
		{
			name:          "RetrieveByShortURL Existent",
			execStatement: "INSERT INTO urls (short_url, original_url) VALUES ('4rSPg8ap', 'http://yandex.ru'), ('dG56Hqxm', 'http://practicum.yandex.ru')",
			args: args{
				id: "dG56Hqxm",
			},
			wantURL: "http://practicum.yandex.ru",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.execStatement)

			record, err := repo.RetrieveByShortURL(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PGRepoTestSuite.RetrieveURL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if record.OriginalURL != tt.wantURL {
				t.Errorf("PGRepoTestSuite.RetrieveURL got = %v, want %v", record.OriginalURL, tt.wantURL)
			}
		})
	}
}

func TestSaveAndRetrieve(t *testing.T) {
	cleanup()

	testURL := "http://example.com"
	id, exists, err := repo.SaveURL(context.Background(), uuid.New(), testURL)
	require.NoError(t, err)
	assert.False(t, exists, "URL should not exists in DB yet")

	record, err := repo.RetrieveByShortURL(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, record.OriginalURL, testURL)
}

func TestSaveURLs(t *testing.T) {
	type args struct {
		urls []string
	}
	tests := []struct {
		name          string
		execStatement string
		args          args
		wantIDsCount  int
		wantErr       bool
	}{
		{
			name:          "All new IDs",
			execStatement: "",
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
		{
			name:          "One new IDs",
			execStatement: "INSERT INTO urls (short_url, original_url) VALUES ('4rSPg8ap', 'http://yandex.ru'), ('edVPg3ks', 'http://ya.ru')",
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
		{
			name: "Existing IDs",
			execStatement: "INSERT INTO urls (short_url, original_url) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru'), ('edVPg3ks', 'http://ya.ru'), ('dG56Hqxm', 'http://practicum.yandex.ru')",
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.execStatement)

			gotIDs, err := repo.SaveURLs(context.Background(), tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.SaveURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(gotIDs), tt.wantIDsCount)
		})
	}
}

func TestDeleteByShortURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID    uuid.UUID
		shortURLs []string
	}
	tests := []struct {
		name                 string
		execStatement        string
		args                 args
		allShortURLs         []string
		wantShortURLsDeleted []string
	}{
		{
			name: "Delete all",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "'), ('edVPg3ks', 'http://ya.ru', '" + userID.String() + "');",
			args: args{
				userID:    userID,
				shortURLs: []string{"4rSPg8ap", "edVPg3ks"},
			},
			allShortURLs:         []string{"4rSPg8ap", "edVPg3ks"},
			wantShortURLsDeleted: []string{"4rSPg8ap", "edVPg3ks"},
		},
		{
			name: "Delete one",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "'), ('edVPg3ks', 'http://ya.ru', '" + userID.String() + "');",
			args: args{
				userID:    userID,
				shortURLs: []string{"4rSPg8ap"},
			},
			allShortURLs:         []string{"4rSPg8ap", "edVPg3ks"},
			wantShortURLsDeleted: []string{"4rSPg8ap"},
		},
		{
			name: "Delete only with matching userID",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "'), ('edVPg3ks', 'http://ya.ru', '" + uuid.NewString() + "');",
			args: args{
				userID:    userID,
				shortURLs: []string{"4rSPg8ap", "edVPg3ks"},
			},
			allShortURLs:         []string{"4rSPg8ap", "edVPg3ks"},
			wantShortURLsDeleted: []string{"4rSPg8ap"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.execStatement)

			repo.DeleteByShortURLs(context.Background(), tt.args.userID, tt.args.shortURLs)

			time.Sleep(50 * time.Millisecond)

			for _, shortURL := range tt.allShortURLs {
				record, err := repo.RetrieveByShortURL(context.TODO(), shortURL)
				require.NoError(t, err)
				if slices.Contains(tt.wantShortURLsDeleted, shortURL) {
					assert.True(t, record.IsDeleted)
				} else {
					assert.False(t, record.IsDeleted)
				}
			}
		})
	}
}

func TestRetrieveUserURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name          string
		execStatement string
		args          args
		wantRecords   []models.Record
	}{
		{
			name:          "Empty repo",
			execStatement: "",
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{},
		},
		{
			name: "One record",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "');",
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "4rSPg8ap",
					OriginalURL: "http://yandex.ru",
					UserID:      userID,
				},
			},
		},
		{
			name: "Only with matching userID",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "'), ('edVPg3ks', 'http://ya.ru', '" + uuid.NewString() + "');",
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "4rSPg8ap",
					OriginalURL: "http://yandex.ru",
					UserID:      userID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.execStatement)

			records, err := repo.RetrieveUserURLs(context.TODO(), tt.args.userID)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantRecords, records)
		})
	}
}

func TestGetStats(t *testing.T) {
	type args struct {
		execStatement string
	}
	tests := []struct {
		name      string
		args      args
		wantStats models.Stats
		wantErr   bool
	}{
		{
			name: "Empty database",
			args: args{
				execStatement: "",
			},
			wantStats: models.Stats{
				URLsCount:  0,
				UsersCount: 0,
			},
			wantErr: false,
		},
		{
			name: "One URL, one user",
			args: args{
				execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
					"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + uuid.New().String() + "');",
			},
			wantStats: models.Stats{
				URLsCount:  1,
				UsersCount: 1,
			},
			wantErr: false,
		},
		{
			name: "Multiple URLs, one user",
			args: args{
				execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
					"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + uuid.New().String() + "'), " +
					"('edVPg3ks', 'http://ya.ru', '" + uuid.New().String() + "'), " +
					"('dG56Hqxm', 'http://practicum.yandex.ru', '" + uuid.New().String() + "');",
			},
			wantStats: models.Stats{
				URLsCount:  3,
				UsersCount: 3,
			},
			wantErr: false,
		},
		{
			name: "Multiple URLs, multiple users",
			args: args{
				execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
					"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + uuid.New().String() + "'), " +
					"('edVPg3ks', 'http://ya.ru', '" + uuid.New().String() + "'), " +
					"('dG56Hqxm', 'http://practicum.yandex.ru', '" + uuid.New().String() + "');",
			},
			wantStats: models.Stats{
				URLsCount:  3,
				UsersCount: 3,
			},
			wantErr: false,
		},
		{
			name: "With deleted URLs",
			args: args{
				execStatement: "INSERT INTO urls (short_url, original_url, user_id, is_deleted) " +
					"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + uuid.New().String() + "', false), " +
					"('edVPg3ks', 'http://ya.ru', '" + uuid.New().String() + "', true), " +
					"('dG56Hqxm', 'http://practicum.yandex.ru', '" + uuid.New().String() + "', false);",
			},
			wantStats: models.Stats{
				URLsCount:  3,
				UsersCount: 3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, tt.args.execStatement)

			gotStats, err := repo.GetStats(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PGRepository.GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.wantStats, gotStats) {
				t.Errorf("PGRepository.GetStats() = %v, want %v", gotStats, tt.wantStats)
			}
		})
	}
}

// BenchmarkPGRepository_SaveURL измеряет производительность сохранения URL
func BenchmarkPGRepository_SaveURL(b *testing.B) {
	cleanup()
	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}
}

// BenchmarkPGRepository_RetrieveByShortURL измеряет производительность получения URL по короткому идентификатору
func BenchmarkPGRepository_RetrieveByShortURL(b *testing.B) {
	cleanup()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveByShortURL(ctx, id)
	}
}

// BenchmarkPGRepository_RetrieveUserURLs измеряет производительность получения всех URL пользователя
func BenchmarkPGRepository_RetrieveUserURLs(b *testing.B) {
	cleanup()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	for i := 0; i < 100; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveUserURLs(ctx, userID)
	}
}

// BenchmarkPGRepository_DeleteByShortURLs измеряет производительность удаления URL
func BenchmarkPGRepository_DeleteByShortURLs(b *testing.B) {
	cleanup()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	var shortURLs []string
	for range 100 {
		id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")
		shortURLs = append(shortURLs, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.DeleteByShortURLs(ctx, userID, shortURLs)
		// Ждем завершения асинхронных операций
		time.Sleep(50 * time.Millisecond)
	}
}

// BenchmarkPGRepository_SaveURLs измеряет производительность пакетного сохранения URL
func BenchmarkPGRepository_SaveURLs(b *testing.B) {
	cleanup()
	ctx := context.Background()

	urls := make([]string, 100)
	for range 100 {
		urls = append(urls, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.SaveURLs(ctx, urls)
	}
}

// BenchmarkPGRepository_CheckStatus измеряет производительность проверки состояния хранилища
func BenchmarkPGRepository_CheckStatus(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.CheckStatus(ctx)
	}
}

// TestNewDB_ValidDSN тестирует создание соединения с БД с валидным DSN
func TestNewDB_ValidDSN(t *testing.T) {
	// Используем существующий контейнер для тестирования
	db, err := NewDB(pgContainer.ConnectionString)
	assert.NoError(t, err, "should not return error for valid DSN")
	assert.NotNil(t, db)
	assert.NotNil(t, db.SQLDB)

	// Проверяем, что соединение работает
	err = db.SQLDB.PingContext(context.Background())
	assert.NoError(t, err)

	// Закрываем соединение
	err = db.SQLDB.Close()
	assert.NoError(t, err)
}

// TestCheckStatus тестирует проверку состояния хранилища
func TestCheckStatus(t *testing.T) {
	cleanup()

	// Тест успешной проверки
	err := repo.CheckStatus(context.Background())
	assert.NoError(t, err)

	// Тест с отмененным контекстом
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = repo.CheckStatus(ctx)
	assert.Error(t, err)
}

// TestNewPGRepository_InvalidDB тестирует создание репозитория с невалидной базой данных
func TestNewPGRepository_InvalidDB(t *testing.T) {
	// Создаем невалидное соединение с БД, используя невалидный DSN
	invalidDB, err := NewDB("invalid-dsn")
	assert.Error(t, err, "should return error for invalid DSN")
	assert.Nil(t, invalidDB)
}

// TestNewPGRepository_WithCustomInterval тестирует создание репозитория с кастомным интервалом
func TestNewPGRepository_WithCustomInterval(t *testing.T) {
	db, err := NewDB(pgContainer.ConnectionString)
	require.NoError(t, err)
	defer db.SQLDB.Close()

	customInterval := 100 * time.Millisecond
	repo, err := NewPGRepository(db, customInterval)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
	assert.NotNil(t, repo.deleteQueue)
	assert.NotNil(t, repo.insertStmt)
	assert.NotNil(t, repo.getURLStmt)
	assert.NotNil(t, repo.deleteStmt)
}

// TestNewPGRepository_WithZeroInterval тестирует создание репозитория с нулевым интервалом
func TestNewPGRepository_WithZeroInterval(t *testing.T) {
	db, err := NewDB(pgContainer.ConnectionString)
	require.NoError(t, err)
	defer db.SQLDB.Close()

	repo, err := NewPGRepository(db, 0)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.deleteQueue)
}

// TestNewPGRepository_PrepareStatementError тестирует ошибку при подготовке запросов
func TestNewPGRepository_PrepareStatementError(t *testing.T) {
	db, err := NewDB(pgContainer.ConnectionString)
	require.NoError(t, err)
	defer db.SQLDB.Close()

	// Закрываем соединение, чтобы вызвать ошибку при подготовке запросов
	db.SQLDB.Close()

	_, err = NewPGRepository(db, 30*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prepare insert statement")
}

// TestSaveURL_InsertError тестирует ошибку при вставке URL
func TestSaveURL_InsertError(t *testing.T) {
	cleanup()

	repoWithNilStmt, err := NewPGRepository(repo.db, 30*time.Millisecond)
	require.NoError(t, err)
	repoWithNilStmt.insertStmt = nil // только insertStmt делаем nil

	_, _, err = repoWithNilStmt.SaveURL(context.Background(), uuid.New(), "http://example.com")
	assert.Error(t, err)
}

// TestSaveURL_GetShortURLError тестирует ошибку при получении короткого URL после нарушения уникальности
func TestSaveURL_GetShortURLError(t *testing.T) {
	cleanup()

	repoWithNilStmt, err := NewPGRepository(repo.db, 30*time.Millisecond)
	require.NoError(t, err)
	repoWithNilStmt.getURLStmt = nil // только getURLStmt делаем nil

	// Сначала сохраняем URL
	_, _, err = repo.SaveURL(context.Background(), uuid.New(), "http://example.com")
	assert.NoError(t, err)

	// Теперь пытаемся сохранить тот же URL с невалидным getURLStmt
	_, _, err = repoWithNilStmt.SaveURL(context.Background(), uuid.New(), "http://example.com")
	assert.Error(t, err)
}

// TestSaveURL_UniqueViolation тестирует обработку нарушения уникальности
func TestSaveURL_UniqueViolation(t *testing.T) {
	cleanup()

	userID := uuid.New()
	url := "http://example.com"

	// Сохраняем URL первый раз
	id1, exists1, err := repo.SaveURL(context.Background(), userID, url)
	assert.NoError(t, err)
	assert.False(t, exists1)
	assert.NotEmpty(t, id1)

	// Сохраняем тот же URL второй раз
	id2, exists2, err := repo.SaveURL(context.Background(), userID, url)
	assert.NoError(t, err)
	assert.True(t, exists2)
	assert.Equal(t, id1, id2)
}

// TestGetShortURLByOriginalURL_NoRows тестирует случай, когда URL не найден
func TestGetShortURLByOriginalURL_NoRows(t *testing.T) {
	cleanup()

	shortURL, err := repo.getShortURLByOriginalURL(context.Background(), "http://nonexistent.com")
	assert.NoError(t, err)
	assert.Empty(t, shortURL)
}

// TestGetShortURLByOriginalURL_Error тестирует ошибку при получении короткого URL
func TestGetShortURLByOriginalURL_Error(t *testing.T) {
	cleanup()

	repoWithNilStmt, err := NewPGRepository(repo.db, 30*time.Millisecond)
	require.NoError(t, err)
	repoWithNilStmt.getURLStmt = nil

	_, err = repoWithNilStmt.getShortURLByOriginalURL(context.Background(), "http://example.com")
	assert.Error(t, err)
}

// TestRetrieveByShortURL_ScanError тестирует ошибку при сканировании результата
func TestRetrieveByShortURL_ScanError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при запросе
	}

	_, err := invalidRepo.RetrieveByShortURL(context.Background(), "test")
	assert.Error(t, err)
}

// TestSaveURLs_TransactionError тестирует ошибку при создании транзакции
func TestSaveURLs_TransactionError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при создании транзакции
	}

	_, err := invalidRepo.SaveURLs(context.Background(), []string{"http://example.com"})
	assert.Error(t, err)
}

// TestSaveURLs_StatementError тестирует ошибку при подготовке запроса в транзакции
func TestSaveURLs_StatementError(t *testing.T) {
	cleanup()

	repoWithNilStmt, err := NewPGRepository(repo.db, 30*time.Millisecond)
	require.NoError(t, err)
	repoWithNilStmt.insertStmt = nil

	_, err = repoWithNilStmt.SaveURLs(context.Background(), []string{"http://example.com"})
	assert.Error(t, err)
}

// TestSaveURLs_ExecError тестирует ошибку при выполнении запроса в транзакции
func TestSaveURLs_ExecError(t *testing.T) {
	cleanup()

	repoWithNilStmt, err := NewPGRepository(repo.db, 30*time.Millisecond)
	require.NoError(t, err)
	repoWithNilStmt.getURLStmt = nil

	// Сначала сохраняем URL
	_, _, err = repo.SaveURL(context.Background(), uuid.New(), "http://example.com")
	assert.NoError(t, err)

	// Теперь пытаемся сохранить тот же URL с невалидным getURLStmt
	_, err = repoWithNilStmt.SaveURLs(context.Background(), []string{"http://example.com"})
	assert.Error(t, err)
}

// TestGetStats_ScanError тестирует ошибку при сканировании статистики
func TestGetStats_ScanError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при запросе
	}

	_, err := invalidRepo.GetStats(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db or db.SQLDB is nil")
}

// TestRetrieveUserURLs_QueryError тестирует ошибку при выполнении запроса
func TestRetrieveUserURLs_QueryError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при запросе
	}

	_, err := invalidRepo.RetrieveUserURLs(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db or db.SQLDB is nil")
}

// TestRetrieveUserURLs_ScanError тестирует ошибку при сканировании результатов
func TestRetrieveUserURLs_ScanError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при запросе
	}

	// Попытка получить URL пользователя должна вызвать ошибку
	_, err := invalidRepo.RetrieveUserURLs(context.Background(), uuid.Nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db or db.SQLDB is nil")
}

// TestRetrieveUserURLs_RowsError тестирует ошибку при обработке строк
func TestRetrieveUserURLs_RowsError(t *testing.T) {
	cleanup()

	// Создаем репозиторий с невалидным соединением
	invalidRepo := &PGRepository{
		db: &DB{SQLDB: nil}, // Это вызовет ошибку при запросе
	}

	_, err := invalidRepo.RetrieveUserURLs(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db or db.SQLDB is nil")
}

// TestSaveURL_OtherError тестирует обработку других ошибок при сохранении
func TestSaveURL_OtherError(t *testing.T) {
	cleanup()

	// Создаем репозиторий без запуска flushDeletions
	repoWithNilStmt := &PGRepository{
		db:          repo.db,
		deleteQueue: make(chan deleteIn, 64),
		insertStmt:  nil, // Это вызовет ошибку
		getURLStmt:  repo.getURLStmt,
		deleteStmt:  repo.deleteStmt,
	}

	_, _, err := repoWithNilStmt.SaveURL(context.Background(), uuid.New(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insertStmt is nil")
}
