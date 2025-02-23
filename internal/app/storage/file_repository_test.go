package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_ReadFromFile(t *testing.T) {
	t.Run("Data from file", func(t *testing.T) {
		fpath := "./test/test_data.txt"
		frepo, err := NewFileRepository(fpath)
		require.NoError(t, err)

		var want = []URLRecord{
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
				{UUID: "1", ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"},
				{UUID: "2", ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"},
				{UUID: "3", ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"},
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
			gotID, gotExists, err := frepo.SaveURL(tt.args.url)
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

func TestFileRepository_RetrieveURL(t *testing.T) {
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
				{UUID: "1", ShortURL: "4rSPg8ap", OriginalURL: "http://yandex.ru"},
				{UUID: "2", ShortURL: "edVPg3ks", OriginalURL: "http://ya.ru"},
				{UUID: "3", ShortURL: "dG56Hqxm", OriginalURL: "http://practicum.yandex.ru"},
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
			gotURL, err := frepo.RetrieveURL(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileRepository.RetrieveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("FileRepository.RetrieveURL() = %v, want %v", gotURL, tt.wantURL)
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
	id, _, _ := frepo.SaveURL(testURL)

	frepo2, err := NewFileRepository(fpath)
	require.NoError(t, err)

	url, err := frepo2.RetrieveURL(id)
	if err != nil {
		t.Errorf("Got unexpected error %s", err.Error())
		return
	}
	if url != testURL {
		t.Errorf("Expected: %s, got: %s", testURL, url)
		return
	}
	os.Remove(fpath)
}
