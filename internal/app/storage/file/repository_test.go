package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iubondar/url-shortener/internal/app/models"
)

func TestFileRepository_ReadFromFile(t *testing.T) {
	t.Run("Data from file", func(t *testing.T) {
		fpath := "./test/test_data.txt"
		frepo, err := NewFileRepository(fpath)
		require.NoError(t, err)

		var want = []URLRecord{
			{UUID: "1", Record: models.Record{ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"}},
			{UUID: "2", Record: models.Record{ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"}},
			{UUID: "3", Record: models.Record{ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"}},
		}

		assert.ElementsMatch(t, want, frepo.records)
	})

	t.Run("Empty file", func(t *testing.T) {
		fpath := os.TempDir() + "frepo_empty_file"
		frepo, err := NewFileRepository(fpath)
		require.NoError(t, err)

		assert.Equal(t, len(frepo.records), 0)

		os.Remove(fpath)
	})

	t.Run("Empty file with nested path", func(t *testing.T) {
		fpath := filepath.Join(os.TempDir(), "a", "b", "frepo_empty_file.txt")
		frepo, err := NewFileRepository(fpath)
		require.NoError(t, err)

		assert.Equal(t, len(frepo.records), 0)

		os.Remove(fpath)
	})
}

func TestFileRepository_SaveURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		records    []URLRecord
		args       args
		wantID     bool
		wantExists bool
		wantErr    bool
	}{
		{
			name:    "Non-existent",
			records: []URLRecord{},
			args: args{
				url: "http://example.com",
			},
			wantID:     true,
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "Existent",
			records: []URLRecord{
				{UUID: "1", Record: models.Record{ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"}},
				{UUID: "2", Record: models.Record{ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"}},
				{UUID: "3", Record: models.Record{ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"}},
			},
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
			fpath := os.TempDir() + "frepo_save_url_tmp"
			frepo := FileRepository{
				fPath:   fpath,
				records: tt.records,
			}
			userID := uuid.New()
			gotID, gotExists, err := frepo.SaveURL(context.Background(), userID, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.SaveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantID && len(gotID) == 0 {
				t.Error("FileRepository.SaveURL() received empty id", gotID, tt.wantID)
			}
			if !tt.wantID && len(gotID) != 0 {
				t.Error("FileRepository.SaveURL() received unexpected id", gotID, tt.wantID)
			}
			if gotExists != tt.wantExists {
				t.Errorf("FileRepository.SaveURL() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
			os.Remove(fpath)
		})
	}
}

func TestFileRepository_RetrieveByShortURL(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		records []URLRecord
		args    args
		wantURL string
		wantErr bool
	}{
		{
			name:    "Non-existent",
			records: []URLRecord{},
			args: args{
				id: "123",
			},
			wantURL: "",
			wantErr: true,
		},
		{
			name: "Existent",
			records: []URLRecord{
				{UUID: "1", Record: models.Record{ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"}},
				{UUID: "2", Record: models.Record{ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"}},
				{UUID: "3", Record: models.Record{ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"}},
			},
			args: args{
				id: "dG56Hqxm",
			},
			wantURL: "http://practicum.yandex.ru",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fpath := os.TempDir() + "frepo_save_url_tmp"
			frepo := FileRepository{
				fPath:   fpath,
				records: tt.records,
			}
			record, err := frepo.RetrieveByShortURL(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.RetrieveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if record.OriginalURL != tt.wantURL {
				t.Errorf("FileRepository.RetrieveURL() = %v, want %v", record.OriginalURL, tt.wantURL)
			}
			os.Remove(fpath)
		})
	}
}

func TestFileRepository_SaveAndRetrieve(t *testing.T) {
	fpath := os.TempDir() + "frepo_save_url_tmp"
	frepo, err := NewFileRepository(fpath)
	require.NoError(t, err)
	testURL := "http://example.com"
	id, _, _ := frepo.SaveURL(context.Background(), uuid.New(), testURL)

	frepo2, err := NewFileRepository(fpath)
	require.NoError(t, err)

	record, err := frepo2.RetrieveByShortURL(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, testURL, record.OriginalURL)

	os.Remove(fpath)
}

func TestFileRepository_SaveURLs(t *testing.T) {
	type fields struct {
		records []URLRecord
	}
	type args struct {
		urls []string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantIDsCount int
		wantErr      bool
	}{
		{
			name: "All new IDs",
			fields: fields{
				records: []URLRecord{},
			},
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
		{
			name: "One new IDs",
			fields: fields{
				records: []URLRecord{
					{UUID: "1", Record: models.Record{ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"}},
					{UUID: "2", Record: models.Record{ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"}},
				},
			},
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
		{
			name: "Existing IDs",
			fields: fields{
				records: []URLRecord{
					{UUID: "1", Record: models.Record{ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"}},
					{UUID: "2", Record: models.Record{ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"}},
					{UUID: "3", Record: models.Record{ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"}},
				},
			},
			args: args{
				urls: []string{"http://yandex.ru", "http://ya.ru", "http://practicum.yandex.ru"},
			},
			wantIDsCount: 3,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fpath := os.TempDir() + "frepo_save_url_tmp"
			frepo := FileRepository{
				fPath:   fpath,
				records: tt.fields.records,
			}
			gotIDs, err := frepo.SaveURLs(context.Background(), tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.SaveURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(gotIDs), tt.wantIDsCount)
			os.Remove(fpath)
		})
	}
}

func TestFileRepository_DeleteByShortURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID    uuid.UUID
		shortURLs []string
	}
	tests := []struct {
		name        string
		records     []URLRecord
		args        args
		wantRecords []URLRecord
	}{
		{
			name:    "Empty repo",
			records: []URLRecord{},
			args: args{
				userID:    userID,
				shortURLs: []string{"hsgdbbn"},
			},
			wantRecords: []URLRecord{},
		},
		{
			name: "One record - deleted successfully",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"123"},
			},
			wantRecords: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
			},
		},
		{
			name: "UserID not match",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				userID:    uuid.New(),
				shortURLs: []string{"123"},
			},
			wantRecords: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
			},
		},
		{
			name: "Delete some",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "456",
						OriginalURL: "http://ya.ru",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "789",
						OriginalURL: "http://avito.ru",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"456"},
			},
			wantRecords: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "456",
						OriginalURL: "http://ya.ru",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "789",
						OriginalURL: "http://avito.ru",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
			},
		},
		{
			name: "Delete all",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "456",
						OriginalURL: "http://ya.ru",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "789",
						OriginalURL: "http://avito.ru",
						UserID:      userID,
						IsDeleted:   false,
					},
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"123", "456", "789"},
			},
			wantRecords: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "456",
						OriginalURL: "http://ya.ru",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
				{
					Record: models.Record{
						ShortURL:    "789",
						OriginalURL: "http://avito.ru",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fpath := os.TempDir() + "frepo_save_url_tmp"
			frepo := FileRepository{
				fPath:   fpath,
				records: tt.records,
			}

			frepo.DeleteByShortURLs(context.Background(), tt.args.userID, tt.args.shortURLs)

			assert.ElementsMatch(t, tt.wantRecords, frepo.records)
			os.Remove(fpath)
		})
	}
}

func TestFileRepository_RetrieveUserURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name        string
		records     []URLRecord
		args        args
		wantRecords []models.Record
	}{
		{
			name:    "Empty repo",
			records: []URLRecord{},
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{},
		},
		{
			name: "One record",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
			},
		},
		{
			name: "UserID not match",
			records: []URLRecord{
				{
					Record: models.Record{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				userID: uuid.New(),
			},
			wantRecords: []models.Record{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fpath := os.TempDir() + "frepo_save_url_tmp"
			frepo := FileRepository{
				fPath:   fpath,
				records: tt.records,
			}

			records, err := frepo.RetrieveUserURLs(context.Background(), tt.args.userID)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantRecords, records)
			os.Remove(fpath)
		})
	}
}

// setupTestFile создает временный файл для тестов
func setupTestFile(t *testing.B) string {
	tempFile := filepath.Join(os.TempDir(), "benchmark_urls.json")
	t.Cleanup(func() {
		os.Remove(tempFile)
	})
	return tempFile
}

// BenchmarkFileRepository_SaveURL измеряет производительность сохранения URL
func BenchmarkFileRepository_SaveURL(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}
}

// BenchmarkFileRepository_RetrieveByShortURL измеряет производительность получения URL по короткому идентификатору
func BenchmarkFileRepository_RetrieveByShortURL(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveByShortURL(ctx, id)
	}
}

// BenchmarkFileRepository_RetrieveUserURLs измеряет производительность получения всех URL пользователя
func BenchmarkFileRepository_RetrieveUserURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	for range 100 {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveUserURLs(ctx, userID)
	}
}

// BenchmarkFileRepository_DeleteByShortURLs измеряет производительность удаления URL
func BenchmarkFileRepository_DeleteByShortURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	var shortURLs []string
	for i := 0; i < 100; i++ {
		id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")
		shortURLs = append(shortURLs, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.DeleteByShortURLs(ctx, userID, shortURLs)
	}
}

// BenchmarkFileRepository_SaveURLs измеряет производительность пакетного сохранения URL
func BenchmarkFileRepository_SaveURLs(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

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

// BenchmarkFileRepository_CheckStatus измеряет производительность проверки состояния хранилища
func BenchmarkFileRepository_CheckStatus(b *testing.B) {
	tempFile := setupTestFile(b)
	repo, err := NewFileRepository(tempFile)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.CheckStatus(ctx)
	}
}
