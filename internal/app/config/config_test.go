package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type envVars struct {
	SERVER_ADDRESS string
	BASE_URL       string
}

func TestConfig_Load(t *testing.T) {

	tests := []struct {
		name    string
		args    []string
		envVars envVars
		want    Config
	}{
		{
			name:    "Defaults",
			args:    nil,
			envVars: envVars{SERVER_ADDRESS: "", BASE_URL: ""},
			want:    Config{ServerAddress: "localhost:8080", BaseURLAddress: "localhost:8080"},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000"},
			envVars: envVars{SERVER_ADDRESS: "", BASE_URL: ""},
			want:    Config{ServerAddress: "localhost:8888", BaseURLAddress: "localhost:8000"},
		},
		{
			name:    "Override with envs",
			args:    []string{"-a", "localhost:8888", "-b", "localhost:8000"},
			envVars: envVars{SERVER_ADDRESS: "localhost:8800", BASE_URL: "localhost:8808"},
			want:    Config{ServerAddress: "localhost:8800", BaseURLAddress: "localhost:8808"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_ADDRESS", tt.envVars.SERVER_ADDRESS)
			t.Setenv("BASE_URL", tt.envVars.BASE_URL)

			var c Config
			c.Load("Test", tt.args)

			assert.Equal(t, tt.want, c)
		})
	}
}
