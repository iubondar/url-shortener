package simple

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iubondar/url-shortener/internal/app/models"
)

func TestSimpleRepository_SaveURL(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []models.Record
	}
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantID     bool
		wantExists bool
		wantErr    bool
	}{
		{
			name: "Non-existent",
			fields: fields{
				records: []models.Record{},
			},
			args: args{
				url: "http://example.com",
			},
			wantID:     true,
			wantExists: false,
			wantErr:    false,
		},
		{
			name: "Existent",
			fields: fields{
				records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				url: "http://example.com",
			},
			wantID:     true,
			wantExists: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := SimpleRepository{
				Records: tt.fields.records,
			}
			gotID, gotExists, err := rep.SaveURL(context.Background(), userID, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.SaveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantID && len(gotID) == 0 {
				t.Error("SimpleRepository.SaveURL() received empty id", gotID, tt.wantID)
			}
			if !tt.wantID && len(gotID) != 0 {
				t.Error("SimpleRepository.SaveURL() received unexpected id", gotID, tt.wantID)
			}
			if gotExists != tt.wantExists {
				t.Errorf("SimpleRepository.SaveURL() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestSimpleRepository_RetrieveByShortURL(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []models.Record
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantURL string
		wantErr bool
	}{
		{
			name: "Non-existent",
			fields: fields{
				records: []models.Record{},
			},
			args: args{
				id: "123",
			},
			wantURL: "",
			wantErr: true,
		},
		{
			name: "Existent",
			fields: fields{
				records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				id: "123",
			},
			wantURL: "http://example.com",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := SimpleRepository{
				Records: tt.fields.records,
			}
			record, err := rep.RetrieveByShortURL(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.RetrieveByShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if record.OriginalURL != tt.wantURL {
				t.Errorf("SimpleRepository.RetrieveByShortURL() = %v, want %v", record.OriginalURL, tt.wantURL)
			}
		})
	}
}

func TestSimpleRepository_SaveAndRetrieve(t *testing.T) {
	rep := NewSimpleRepository()
	testURL := "http://example.com"
	ctx := context.Background()
	id, _, _ := rep.SaveURL(ctx, uuid.New(), testURL)
	record, err := rep.RetrieveByShortURL(ctx, id)
	if err != nil {
		t.Errorf("Got unexpected error %s", err.Error())
		return
	}
	if record.OriginalURL != testURL {
		t.Errorf("Expected: %s, got: %s", testURL, record.OriginalURL)
		return
	}
}

func TestSimpleRepository_RetrieveID(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []models.Record
	}
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantID  string
		wantErr bool
	}{
		{
			name: "Non-existent",
			fields: fields{
				records: []models.Record{},
			},
			args: args{
				url: "http://example.com",
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name: "Existent",
			fields: fields{
				records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				url: "http://example.com",
			},
			wantID:  "123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := SimpleRepository{
				Records: tt.fields.records,
			}
			gotID, err := rep.RetrieveID(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.RetrieveID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("SimpleRepository.RetrieveID() = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestSimpleRepository_SaveURLs(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []models.Record
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
				records: []models.Record{},
			},
			args: args{
				urls: []string{"http://example.com", "http://ya.ru"},
			},
			wantIDsCount: 2,
			wantErr:      false,
		},
		{
			name: "One new IDs",
			fields: fields{
				records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
				},
			},
			args: args{
				urls: []string{"http://example.com", "http://ya.ru"},
			},
			wantIDsCount: 2,
			wantErr:      false,
		},
		{
			name: "Existing IDs",
			fields: fields{
				records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: "http://example.com",
						UserID:      userID,
					},
					{
						ShortURL:    "456",
						OriginalURL: "http://ya.ru",
						UserID:      userID,
					},
				},
			},
			args: args{
				urls: []string{"http://example.com", "http://ya.ru"},
			},
			wantIDsCount: 2,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := SimpleRepository{
				Records: tt.fields.records,
			}
			gotIDs, err := repo.SaveURLs(context.Background(), tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.SaveURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(gotIDs), tt.wantIDsCount)
		})
	}
}

func TestSimpleRepository_DeleteByShortURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID    uuid.UUID
		shortURLs []string
	}
	tests := []struct {
		name        string
		records     []models.Record
		args        args
		wantRecords []models.Record
	}{
		{
			name:    "Empty repo",
			records: []models.Record{},
			args: args{
				userID:    userID,
				shortURLs: []string{"hsgdbbn"},
			},
			wantRecords: []models.Record{},
		},
		{
			name: "One record - deleted successfully",
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"123"},
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   true,
				},
			},
		},
		{
			name: "UserID not match",
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
			},
			args: args{
				userID:    uuid.New(),
				shortURLs: []string{"123"},
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   false,
				},
			},
		},
		{
			name: "Delete some",
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   false,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      userID,
					IsDeleted:   false,
				},
				{
					ShortURL:    "789",
					OriginalURL: "http://avito.ru",
					UserID:      userID,
					IsDeleted:   false,
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"456"},
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   false,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      userID,
					IsDeleted:   true,
				},
				{
					ShortURL:    "789",
					OriginalURL: "http://avito.ru",
					UserID:      userID,
					IsDeleted:   false,
				},
			},
		},
		{
			name: "Delete all",
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   false,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      userID,
					IsDeleted:   false,
				},
				{
					ShortURL:    "789",
					OriginalURL: "http://avito.ru",
					UserID:      userID,
					IsDeleted:   false,
				},
			},
			args: args{
				userID:    userID,
				shortURLs: []string{"123", "456", "789"},
			},
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   true,
				},
				{
					ShortURL:    "456",
					OriginalURL: "http://ya.ru",
					UserID:      userID,
					IsDeleted:   true,
				},
				{
					ShortURL:    "789",
					OriginalURL: "http://avito.ru",
					UserID:      userID,
					IsDeleted:   true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := SimpleRepository{
				Records: tt.records,
			}
			repo.DeleteByShortURLs(context.Background(), tt.args.userID, tt.args.shortURLs)

			assert.ElementsMatch(t, tt.wantRecords, repo.Records)
		})
	}
}

func TestSimpleRepository_RetrieveUserURLs(t *testing.T) {
	userID := uuid.New()
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name        string
		records     []models.Record
		args        args
		wantRecords []models.Record
	}{
		{
			name:    "Empty repo",
			records: []models.Record{},
			args: args{
				userID: userID,
			},
			wantRecords: []models.Record{},
		},
		{
			name: "One record",
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
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
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
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
			repo := SimpleRepository{
				Records: tt.records,
			}
			records, err := repo.RetrieveUserURLs(context.Background(), tt.args.userID)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantRecords, records)
		})
	}
}

// BenchmarkSimpleRepository_SaveURL измеряет производительность сохранения URL
func BenchmarkSimpleRepository_SaveURL(b *testing.B) {
	repo := NewSimpleRepository()
	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.SaveURL(ctx, userID, "http://example.com")
	}
}

// BenchmarkSimpleRepository_RetrieveByShortURL измеряет производительность получения URL по короткому идентификатору
func BenchmarkSimpleRepository_RetrieveByShortURL(b *testing.B) {
	repo := NewSimpleRepository()
	userID := uuid.New()
	ctx := context.Background()

	// Подготовка данных
	id, _, _ := repo.SaveURL(ctx, userID, "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.RetrieveByShortURL(ctx, id)
	}
}

// BenchmarkSimpleRepository_RetrieveUserURLs измеряет производительность получения всех URL пользователя
func BenchmarkSimpleRepository_RetrieveUserURLs(b *testing.B) {
	repo := NewSimpleRepository()
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

// BenchmarkSimpleRepository_DeleteByShortURLs измеряет производительность удаления URL
func BenchmarkSimpleRepository_DeleteByShortURLs(b *testing.B) {
	repo := NewSimpleRepository()
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
	}
}

// BenchmarkSimpleRepository_SaveURLs измеряет производительность пакетного сохранения URL
func BenchmarkSimpleRepository_SaveURLs(b *testing.B) {
	repo := NewSimpleRepository()
	ctx := context.Background()

	urls := make([]string, 100)
	for i := range 100 {
		urls[i] = "http://example.com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.SaveURLs(ctx, urls)
	}
}
