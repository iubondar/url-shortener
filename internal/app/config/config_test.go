package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {

	tests := []struct {
		name    string
		args    []string
		envVars Config
		want    Config
	}{
		{
			name:    "Defaults",
			args:    nil,
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: ""},
			want:    Config{ServerAddress: "localhost:8080", BaseURLAddress: "localhost:8080", FileStoragePath: "./storage/storage.txt"},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt"},
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: ""},
			want:    Config{ServerAddress: "localhost:8888", BaseURLAddress: "localhost:8000", FileStoragePath: "st/base.txt"},
		},
		{
			name:    "Override with envs",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt"},
			envVars: Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt"},
			want:    Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.envVars.ServerAddress)
			t.Setenv("BASE_URL", tt.envVars.BaseURLAddress)
			t.Setenv("FILE_STORAGE_PATH", tt.envVars.FileStoragePath)

			var c Config
			c.Load("Test", tt.args)

			assert.Equal(t, tt.want, c)
		})
	}
}
