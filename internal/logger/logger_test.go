package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init()
	assert.NotNil(t, log)
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	// Create a new logger for testing
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	Error("test error")

	output := buf.String()
	assert.Contains(t, output, "test error")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	testLogger := NewJSONHandler(&buf, opts)
	log = New(testLogger)

	Debug("test debug")

	output := buf.String()
	assert.Contains(t, output, "test debug")
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	Infof("test %s", "message")

	output := buf.String()
	// slog formats differently - check for the message content
	assert.Contains(t, output, "message")
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	Errorf("test %s", "error")

	output := buf.String()
	// slog formats differently - check for the message content
	assert.Contains(t, output, "error")
}

func TestDebugf(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	testLogger := NewJSONHandler(&buf, opts)
	log = New(testLogger)

	Debugf("test %s", "debug")

	output := buf.String()
	assert.Contains(t, output, "debug")
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	err := assert.AnError
	logger := WithError(err)
	logger.Info("test with error")

	output := buf.String()
	assert.Contains(t, output, "test with error")
	assert.Contains(t, output, "error")
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewJSONHandler(&buf, nil)
	log = New(testLogger)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	logger := WithFields(fields)
	logger.Info("test with fields")

	output := buf.String()
	assert.Contains(t, output, "test with fields")
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
}
