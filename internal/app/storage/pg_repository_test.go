package storage

import (
	"context"
	"database/sql"
	"log"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"

	"github.com/iubondar/url-shortener/internal/app/storage/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PGRepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	repo        *PGRepository
	ctx         context.Context
}

func (suite *PGRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	db, err := sql.Open("pgx", suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	goose.SetDialect("postgres")
	err = goose.Up(db, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	pgRepo, err := NewPGRepository(db, 30*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	suite.repo = pgRepo
}

func (suite *PGRepoTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *PGRepoTestSuite) SetupTest() {
	err := suite.clearUrlsTable()
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *PGRepoTestSuite) clearUrlsTable() error {
	_, err := suite.repo.db.ExecContext(suite.ctx, "TRUNCATE TABLE urls;")
	return err
}

func setupSeparateTest(t *testing.T, suite *PGRepoTestSuite, execStatement string) {
	err := suite.clearUrlsTable()
	require.NoError(t, err)

	if len(execStatement) > 0 {
		_, err := suite.repo.db.ExecContext(suite.ctx, execStatement)
		require.NoError(t, err)
	}
}

// Tests are here
func (suite *PGRepoTestSuite) TestSaveURL() {
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
	t := suite.T()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, suite, tt.execStatement)

			gotID, gotExists, err := suite.repo.SaveURL(suite.ctx, userID, tt.args.url)

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

func (suite *PGRepoTestSuite) TestRetrieveByShortURL() {
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
	t := suite.T()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, suite, tt.execStatement)

			record, err := suite.repo.RetrieveByShortURL(context.Background(), tt.args.id)
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

func (suite *PGRepoTestSuite) TestSaveAndRetrieve() {
	t := suite.T()
	err := suite.clearUrlsTable()
	require.NoError(t, err)

	testURL := "http://example.com"
	id, exists, err := suite.repo.SaveURL(suite.ctx, uuid.New(), testURL)
	require.NoError(t, err)
	assert.False(t, exists, "URL should not exists in DB yet")

	record, err := suite.repo.RetrieveByShortURL(suite.ctx, id)
	require.NoError(t, err)
	assert.Equal(t, record.OriginalURL, testURL)
}

func (suite *PGRepoTestSuite) TestSaveURLs() {
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
	t := suite.T()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, suite, tt.execStatement)

			gotIDs, err := suite.repo.SaveURLs(context.Background(), tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.SaveURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(gotIDs), tt.wantIDsCount)
		})
	}
}

func (suite *PGRepoTestSuite) TestDeleteByShortURLs() {
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
	t := suite.T()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, suite, tt.execStatement)

			suite.repo.DeleteByShortURLs(context.Background(), tt.args.userID, tt.args.shortURLs)

			time.Sleep(50 * time.Millisecond)

			for _, shortURL := range tt.allShortURLs {
				record, err := suite.repo.RetrieveByShortURL(context.TODO(), shortURL)
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

func (suite *PGRepoTestSuite) TestRetrieveUserURLs() {
	userID := uuid.New()
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name          string
		execStatement string
		args          args
		wantRecords   []Record
	}{
		{
			name:          "Empty repo",
			execStatement: "",
			args: args{
				userID: userID,
			},
			wantRecords: []Record{},
		},
		{
			name: "One record",
			execStatement: "INSERT INTO urls (short_url, original_url, user_id) " +
				"VALUES ('4rSPg8ap', 'http://yandex.ru', '" + userID.String() + "');",
			args: args{
				userID: userID,
			},
			wantRecords: []Record{
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
			wantRecords: []Record{
				{
					ShortURL:    "4rSPg8ap",
					OriginalURL: "http://yandex.ru",
					UserID:      userID,
				},
			},
		},
	}
	t := suite.T()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSeparateTest(t, suite, tt.execStatement)

			records, err := suite.repo.RetrieveUserURLs(context.TODO(), tt.args.userID)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantRecords, records)
		})
	}
}

// Запуск сьюта тестов
func TestPGRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PGRepoTestSuite))
}
