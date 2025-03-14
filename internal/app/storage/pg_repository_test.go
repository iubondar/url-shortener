package storage

import (
	"context"
	"database/sql"
	"log"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

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

	pgRepo, err := NewPGRepository(suite.ctx, db)
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

// Tests are here
func (suite *PGRepoTestSuite) TestSaveURL() {
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
			name:          "SaveURL Existent",
			execStatement: "INSERT INTO urls (short_url, original_url) VALUES ('4rSPg8ap', 'http://yandex.ru'), ('edVPg3ks', 'http://ya.ru')",
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
			err := suite.clearUrlsTable()
			require.NoError(t, err)

			if len(tt.execStatement) > 0 {
				_, err = suite.repo.db.ExecContext(suite.ctx, tt.execStatement)
				require.NoError(t, err)
			}

			gotID, gotExists, err := suite.repo.SaveURL(suite.ctx, tt.args.url)

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

func (suite *PGRepoTestSuite) TestRetrieveURL() {
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
			name:          "RetrieveURL Non-existent",
			execStatement: "",
			args: args{
				id: "123",
			},
			wantURL: "",
			wantErr: true,
		},
		{
			name:          "RetrieveURL Existent",
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
			err := suite.clearUrlsTable()
			require.NoError(t, err)

			if len(tt.execStatement) > 0 {
				_, err := suite.repo.db.ExecContext(suite.ctx, tt.execStatement)
				require.NoError(t, err)
			}

			gotURL, err := suite.repo.RetrieveURL(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PGRepoTestSuite.RetrieveURL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("PGRepoTestSuite.RetrieveURL got = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func (suite *PGRepoTestSuite) TestSaveAndRetrieve() {
	t := suite.T()
	err := suite.clearUrlsTable()
	require.NoError(t, err)

	testURL := "http://example.com"
	id, exists, err := suite.repo.SaveURL(suite.ctx, testURL)
	require.NoError(t, err)
	assert.False(t, exists, "URL should not exists in DB yet")

	url, err := suite.repo.RetrieveURL(suite.ctx, id)
	require.NoError(t, err)
	assert.Equal(t, url, testURL)
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
			err := suite.clearUrlsTable()
			require.NoError(t, err)

			if len(tt.execStatement) > 0 {
				_, err := suite.repo.db.ExecContext(suite.ctx, tt.execStatement)
				require.NoError(t, err)
			}
			gotIDs, err := suite.repo.SaveURLs(context.Background(), tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.SaveURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, len(gotIDs), tt.wantIDsCount)
		})
	}
}

// Запуск сьюта тестов
func TestCustomerRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PGRepoTestSuite))
}
