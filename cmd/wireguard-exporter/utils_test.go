package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePorts(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "default port",
			input:    DefaultPort,
			expected: ":" + strconv.Itoa(DefaultPort),
		},
		{
			name:     "custom port",
			input:    8080,
			expected: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := parsePort(tt.input)
			assert.NoError(t, err)
			assert.Equalf(t, tt.expected, port, "input=%q, expected=%q, got=%q",
				tt.input, tt.expected, port)
		})
	}
}

func TestInvalidPorts(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
		errMsg   string
	}{
		{
			name:     "port value less than lower limit",
			input:    0,
			expected: "",
			errMsg:   "port must be between 1024 and 49151, got 0",
		},
		{
			name:     "port value greater than upper limit",
			input:    69151,
			expected: "",
			errMsg:   "port must be between 1024 and 49151, got 69151",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := parsePort(tt.input)
			assert.EqualError(t, err, tt.errMsg, "err=%q", err)
			assert.Equalf(t, tt.expected, port, "input=%q, expected=%q, got=%q",
				tt.input, tt.expected, port)
		})
	}
}

func TestGetEnvStr(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("TEST_STR_VAR", "hello")
		assert.Equal(t, "hello", getEnvStr("TEST_STR_VAR", "default"))
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		os.Unsetenv("TEST_STR_VAR_MISSING")
		assert.Equal(t, "default", getEnvStr("TEST_STR_VAR_MISSING", "default"))
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("TEST_INT_VAR", "8080")
		assert.Equal(t, 8080, getEnvInt("TEST_INT_VAR", 9011))
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		os.Unsetenv("TEST_INT_VAR_MISSING")
		assert.Equal(t, 9011, getEnvInt("TEST_INT_VAR_MISSING", 9011))
	})

	t.Run("returns fallback for invalid value", func(t *testing.T) {
		t.Setenv("TEST_INT_VAR_BAD", "notanumber")
		assert.Equal(t, 9011, getEnvInt("TEST_INT_VAR_BAD", 9011))
	})
}
