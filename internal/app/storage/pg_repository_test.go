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

// Tests are here
func (suite *PGRepoTestSuite) TestSaveAndRetrieve() {
	t := suite.T()

	testURL := "http://example.com"
	id, exists, err := suite.repo.SaveURL(suite.ctx, testURL)
	require.NoError(t, err)
	assert.False(t, exists, "URL should not exists in DB yet")

	url, err := suite.repo.RetrieveURL(suite.ctx, id)
	require.NoError(t, err)
	assert.Equal(t, url, testURL)
}

// Запуск сьюта тестов
func TestCustomerRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PGRepoTestSuite))
}
