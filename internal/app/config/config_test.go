package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		envVars map[string]string
		want    Config
	}{
		{
			name:    "Defaults",
			args:    nil,
			envVars: nil,
			want: Config{
				ServerAddress:   defaultAddress,
				BaseURLAddress:  defaultAddress,
				FileStoragePath: defaultStoragePath,
				DatabaseDSN:     defaultDatabaseDSN(),
				EnableHTTPS:     false,
			},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db", "-s"},
			envVars: nil,
			want: Config{
				ServerAddress:   "localhost:8888",
				BaseURLAddress:  "localhost:8000",
				FileStoragePath: "st/base.txt",
				DatabaseDSN:     "host=local user=u password=p dbname=db",
				EnableHTTPS:     true,
			},
		},
		{
			name: "Override with envs",
			args: []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db", "-s"},
			envVars: map[string]string{
				"SERVER_ADDRESS":    "localhost:8800",
				"BASE_URL":          "localhost:8808",
				"FILE_STORAGE_PATH": "./ddd/ttt.txt",
				"DATABASE_DSN":      "dsn",
				"ENABLE_HTTPS":      "false",
			},
			want: Config{
				ServerAddress:   "localhost:8800",
				BaseURLAddress:  "localhost:8808",
				FileStoragePath: "./ddd/ttt.txt",
				DatabaseDSN:     "dsn",
				EnableHTTPS:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем все переменные окружения перед каждым тестом
			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
			os.Unsetenv("FILE_STORAGE_PATH")
			os.Unsetenv("DATABASE_DSN")
			os.Unsetenv("ENABLE_HTTPS")

			// Устанавливаем переменные окружения только если они заданы в тесте
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}
			}

			c, err := NewConfig("Test", tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, c)
		})
	}
}

func TestConfig_EnableHTTPS(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		envVars map[string]string
		want    bool
	}{
		{
			name:    "Neither flag nor env set",
			args:    nil,
			envVars: nil,
			want:    false,
		},
		{
			name:    "Flag set, env not set",
			args:    []string{"-s"},
			envVars: nil,
			want:    true,
		},
		{
			name: "Both flag and env set to true",
			args: []string{"-s"},
			envVars: map[string]string{
				"ENABLE_HTTPS": "true",
			},
			want: true,
		},
		{
			name: "Flag set, env set to false (env has priority)",
			args: []string{"-s"},
			envVars: map[string]string{
				"ENABLE_HTTPS": "false",
			},
			want: false,
		},
		{
			name: "Only env set to true",
			args: nil,
			envVars: map[string]string{
				"ENABLE_HTTPS": "true",
			},
			want: true,
		},
		{
			name: "Only env set to false",
			args: nil,
			envVars: map[string]string{
				"ENABLE_HTTPS": "false",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем переменную окружения перед каждым тестом
			os.Unsetenv("ENABLE_HTTPS")

			// Устанавливаем переменную окружения только если она задана в тесте
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}
			}

			c, err := NewConfig("Test", tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, c.EnableHTTPS)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name        string
		shortConfig string
		longConfig  string
		envVar      string
		want        string
		wantErr     bool
	}{
		{
			name:        "No flags or env",
			shortConfig: "",
			longConfig:  "",
			envVar:      "",
			want:        "",
			wantErr:     false,
		},
		{
			name:        "Short flag",
			shortConfig: "config.json",
			longConfig:  "",
			envVar:      "",
			want:        "config.json",
			wantErr:     false,
		},
		{
			name:        "Long flag",
			shortConfig: "",
			longConfig:  "config.json",
			envVar:      "",
			want:        "config.json",
			wantErr:     false,
		},
		{
			name:        "Both flags error",
			shortConfig: "short.json",
			longConfig:  "long.json",
			envVar:      "",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Env variable",
			shortConfig: "",
			longConfig:  "",
			envVar:      "env_config.json",
			want:        "env_config.json",
			wantErr:     false,
		},
		{
			name:        "Flag overrides env",
			shortConfig: "flag.json",
			longConfig:  "",
			envVar:      "env_config.json",
			want:        "flag.json",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("CONFIG", tt.envVar)
			} else {
				t.Setenv("CONFIG", "")
			}

			got, err := getConfigPath(tt.shortConfig, tt.longConfig)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		envVars map[string]string
		want    Config
	}{
		{
			name:    "Load from file only",
			args:    []string{"-c", "testfiles/test_config.json"},
			envVars: nil,
			want: Config{
				ServerAddress:   "localhost:8080",
				BaseURLAddress:  "http://localhost",
				FileStoragePath: "/path/to/file.db",
				DatabaseDSN:     defaultDatabaseDSN(),
				EnableHTTPS:     true,
			},
		},
		{
			name: "File values overridden by flags",
			args: []string{
				"-c", "testfiles/test_config.json",
				"-a", "localhost:8888",
				"-b", "http://localhost:8888",
				"-f", "custom/path.json",
				"-d", "custom_dsn",
				"-s",
			},
			envVars: nil,
			want: Config{
				ServerAddress:   "localhost:8888",
				BaseURLAddress:  "http://localhost:8888",
				FileStoragePath: "custom/path.json",
				DatabaseDSN:     "custom_dsn",
				EnableHTTPS:     true,
			},
		},
		{
			name: "File values overridden by env",
			args: []string{"-c", "testfiles/test_config.json"},
			envVars: map[string]string{
				"SERVER_ADDRESS":    "localhost:7777",
				"BASE_URL":          "http://localhost:7777",
				"FILE_STORAGE_PATH": "env/path.json",
				"DATABASE_DSN":      "env_dsn",
				"ENABLE_HTTPS":      "false",
			},
			want: Config{
				ServerAddress:   "localhost:7777",
				BaseURLAddress:  "http://localhost:7777",
				FileStoragePath: "env/path.json",
				DatabaseDSN:     "env_dsn",
				EnableHTTPS:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем все переменные окружения перед каждым тестом
			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
			os.Unsetenv("FILE_STORAGE_PATH")
			os.Unsetenv("DATABASE_DSN")
			os.Unsetenv("ENABLE_HTTPS")

			// Устанавливаем переменные окружения только если они заданы в тесте
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}
			}

			// Для отладки: загрузим и проверим содержимое файла
			if tt.name == "Load from file only" {
				fileConfig, err := loadConfigFromFile("testfiles/test_config.json")
				if err != nil {
					t.Fatalf("Failed to load config file: %v", err)
				}
				t.Logf("Loaded from file: %+v", fileConfig)
			}

			c, err := NewConfig("Test", tt.args)
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}
			t.Logf("Final config: %+v", c)
			assert.Equal(t, tt.want, c)
		})
	}
}

func TestConfig_LoadFromFile_Errors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "File not exists",
			args:    []string{"-c", "not_exists.json"},
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			args:    []string{"-c", "testdata/invalid.json"},
			wantErr: true,
		},
	}

	// Создаем файл с невалидным JSON
	err := os.MkdirAll("testdata", 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("testdata")

	invalidJSON := `{"server_address": "localhost:9999", "base_url": "http://localhost:9999", "file_storage_path": "/tmp/storage.json", "database_dsn": "host=test user=test password=test dbname=test", "enable_https": true,}`
	if err := os.WriteFile("testdata/invalid.json", []byte(invalidJSON), 0644); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfig("Test", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
