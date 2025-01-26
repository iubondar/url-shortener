package storage

import (
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
		wantId     bool
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
			wantId:     true,
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
			wantId:     true,
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
			gotId, gotExists, err := rep.SaveURL(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.SaveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantId && len(gotId) == 0 {
				t.Error("SimpleRepository.SaveURL() received empty id", gotId, tt.wantId)
			}
			if !tt.wantId && len(gotId) != 0 {
				t.Error("SimpleRepository.SaveURL() received unexpected id", gotId, tt.wantId)
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
		wantUrl string
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
			wantUrl: "",
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
			wantUrl: "http://example.com",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := SimpleRepository{
				UrlsToIds: tt.fields.urlsToIds,
				IdsToURLs: tt.fields.idsToURLs,
			}
			gotUrl, err := rep.RetrieveURL(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleRepository.RetrieveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUrl != tt.wantUrl {
				t.Errorf("SimpleRepository.RetrieveURL() = %v, want %v", gotUrl, tt.wantUrl)
			}
		})
	}
}

func TestSimpleRepository_SaveAndRetrieve(t *testing.T) {
	rep := NewSimpleRepository()
	testUrl := "http://example.com"
	id, _, _ := rep.SaveURL(testUrl)
	url, err := rep.RetrieveURL(id)
	if err != nil {
		t.Errorf("Got unexpected error %s", err.Error())
		return
	}
	if url != testUrl {
		t.Errorf("Expected: %s, got: %s", testUrl, url)
		return
	}
}
