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

func init() {
	ctx := context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create postgres container: %v", err)
	}

	db, err := NewDB(pgContainer.ConnectionString)
	handleError(err, db, pgContainer, ctx, "Failed to create database connection")

	err = goose.SetDialect("postgres")
	handleError(err, db, pgContainer, ctx, "Failed to set dialect")

	err = goose.Up(db.SQLDB, "./migrations")
	handleError(err, db, pgContainer, ctx, "Failed to run migrations")

	repo, err = NewPGRepository(db, 30*time.Millisecond)
	handleError(err, db, pgContainer, ctx, "Failed to create repository")

	cleanup = func() {
		if repo != nil && repo.db != nil && repo.db.SQLDB != nil {
			_, err := repo.db.SQLDB.ExecContext(context.Background(), "TRUNCATE TABLE urls;")
			if err != nil {
				log.Printf("Failed to clear urls table: %v", err)
			}
		}
	}
}

func TestMain(m *testing.M) {
	// Запускаем все тесты и примеры
	code := m.Run()

	// Очищаем ресурсы после завершения всех тестов
	if repo != nil && repo.db != nil {
		cleanupResources(repo.db, pgContainer, context.Background())
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
	cleanup()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.CheckStatus(ctx)
	}
}
