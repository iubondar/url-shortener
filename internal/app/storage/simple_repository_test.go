package storage

import (
	"context"
	"testing"
)

func TestSimpleRepository_SaveURL(t *testing.T) {
	type fields struct {
		urlsToIds map[string]string
		idsToURLs map[string]string
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				urlsToIds: map[string]string{"http://example.com": "123"},
				idsToURLs: map[string]string{"123": "http://example.com"},
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
				UrlsToIds: tt.fields.urlsToIds,
				IdsToURLs: tt.fields.idsToURLs,
			}
			gotID, gotExists, err := rep.SaveURL(context.Background(), tt.args.url)
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

func TestSimpleRepository_RetrieveURL(t *testing.T) {
	type fields struct {
		urlsToIds map[string]string
		idsToURLs map[string]string
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				urlsToIds: map[string]string{"http://example.com": "123"},
				idsToURLs: map[string]string{"123": "http://example.com"},
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
				UrlsToIds: tt.fields.urlsToIds,
				IdsToURLs: tt.fields.idsToURLs,
			}
			gotURL, err := rep.RetrieveURL(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.RetrieveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("SimpleRepository.RetrieveURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestSimpleRepository_SaveAndRetrieve(t *testing.T) {
	rep := NewSimpleRepository()
	testURL := "http://example.com"
	ctx := context.Background()
	id, _, _ := rep.SaveURL(ctx, testURL)
	url, err := rep.RetrieveURL(ctx, id)
	if err != nil {
		t.Errorf("Got unexpected error %s", err.Error())
		return
	}
	if url != testURL {
		t.Errorf("Expected: %s, got: %s", testURL, url)
		return
	}
}
