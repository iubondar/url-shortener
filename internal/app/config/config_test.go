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
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: "", DatabaseDSN: ""},
			want: Config{
				ServerAddress:   defaultAddress,
				BaseURLAddress:  defaultAddress,
				FileStoragePath: defaultStoragePath,
				DatabaseDSN:     defaultDatabaseDSN(),
			},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db"},
			envVars: Config{ServerAddress: "", BaseURLAddress: "", FileStoragePath: ""},
			want: Config{
				ServerAddress:   "localhost:8888",
				BaseURLAddress:  "localhost:8000",
				FileStoragePath: "st/base.txt",
				DatabaseDSN:     "host=local user=u password=p dbname=db",
			},
		},
		{
			name:    "Override with envs",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000", "-f", "st/base.txt", "-d", "host=local user=u password=p dbname=db"},
			envVars: Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt", DatabaseDSN: "dsn"},
			want:    Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808", FileStoragePath: "./ddd/ttt.txt", DatabaseDSN: "dsn"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.envVars.ServerAddress)
			t.Setenv("BASE_URL", tt.envVars.BaseURLAddress)
			t.Setenv("FILE_STORAGE_PATH", tt.envVars.FileStoragePath)
			t.Setenv("DATABASE_DSN", tt.envVars.DatabaseDSN)

			c, err := NewConfig("Test", tt.args)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, *c)
		})
	}
}
