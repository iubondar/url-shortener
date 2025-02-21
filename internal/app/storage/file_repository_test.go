package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_ReadFromFile(t *testing.T) {
	t.Run("Data from file", func(t *testing.T) {
		fpath := "./test/test_data.txt"
		frepo, err := NewFileRepository(fpath)
		require.NoError(t, err)

		var want []URLRecord = []URLRecord{
			{UUID: "1", ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"},
			{UUID: "2", ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"},
			{UUID: "3", ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"},
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
}
