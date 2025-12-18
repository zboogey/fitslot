package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init()

	assert.NotNil(t, InfoLogger)
	assert.NotNil(t, ErrorLogger)
	assert.NotNil(t, DebugLogger)
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	InfoLogger = log.New(&buf, "INFO: ", log.Lshortfile)

	Info("test message")

	output := buf.String()
	assert.Contains(t, output, "INFO:")
	assert.Contains(t, output, "test message")
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	InfoLogger = log.New(&buf, "INFO: ", log.Lshortfile)

	Infof("test %s %d", "message", 42)

	output := buf.String()
	assert.Contains(t, output, "INFO:")
	assert.Contains(t, output, "test message 42")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	ErrorLogger = log.New(&buf, "ERROR: ", log.Lshortfile)

	Error("error occurred")

	output := buf.String()
	assert.Contains(t, output, "ERROR:")
	assert.Contains(t, output, "error occurred")
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	ErrorLogger = log.New(&buf, "ERROR: ", log.Lshortfile)

	Errorf("error code: %d, message: %s", 500, "internal error")

	output := buf.String()
	assert.Contains(t, output, "ERROR:")
	assert.Contains(t, output, "error code: 500")
	assert.Contains(t, output, "message: internal error")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	DebugLogger = log.New(&buf, "DEBUG: ", log.Lshortfile)

	Debug("debug info")

	output := buf.String()
	assert.Contains(t, output, "DEBUG:")
	assert.Contains(t, output, "debug info")
}

func TestDebugf(t *testing.T) {
	var buf bytes.Buffer
	DebugLogger = log.New(&buf, "DEBUG: ", log.Lshortfile)

	Debugf("variable x = %d", 100)

	output := buf.String()
	assert.Contains(t, output, "DEBUG:")
	assert.Contains(t, output, "variable x = 100")
}

func TestMultipleLoggers(t *testing.T) {
	var infoBuf bytes.Buffer
	var errorBuf bytes.Buffer
	var debugBuf bytes.Buffer

	InfoLogger = log.New(&infoBuf, "INFO: ", 0)
	ErrorLogger = log.New(&errorBuf, "ERROR: ", 0)
	DebugLogger = log.New(&debugBuf, "DEBUG: ", 0)

	Info("info message")
	Error("error message")
	Debug("debug message")

	// Проверяем что каждый логгер пишет в свой буфер
	assert.Contains(t, infoBuf.String(), "info message")
	assert.Contains(t, errorBuf.String(), "error message")
	assert.Contains(t, debugBuf.String(), "debug message")

	// Проверяем что сообщения не попали в другие буферы
	assert.NotContains(t, infoBuf.String(), "error message")
	assert.NotContains(t, errorBuf.String(), "debug message")
}

func TestLoggerPrefix(t *testing.T) {
	var buf bytes.Buffer
	InfoLogger = log.New(&buf, "INFO: ", 0)
	ErrorLogger = log.New(&buf, "ERROR: ", 0)
	DebugLogger = log.New(&buf, "DEBUG: ", 0)

	Info("test1")
	Error("test2")
	Debug("test3")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, 3, len(lines))
	assert.True(t, strings.HasPrefix(lines[0], "INFO:"))
	assert.True(t, strings.HasPrefix(lines[1], "ERROR:"))
	assert.True(t, strings.HasPrefix(lines[2], "DEBUG:"))
}

func TestLoggerNotInitialized(t *testing.T) {
	// Сохраняем текущие логгеры
	oldInfo := InfoLogger
	oldError := ErrorLogger
	oldDebug := DebugLogger

	// Обнуляем логгеры
	InfoLogger = nil
	ErrorLogger = nil
	DebugLogger = nil

	// Восстанавливаем после теста
	defer func() {
		InfoLogger = oldInfo
		ErrorLogger = oldError
		DebugLogger = oldDebug
	}()

	// Проверяем что логгеры nil
	assert.Nil(t, InfoLogger)
	assert.Nil(t, ErrorLogger)
	assert.Nil(t, DebugLogger)

	// После Init они должны быть инициализированы
	Init()

	assert.NotNil(t, InfoLogger)
	assert.NotNil(t, ErrorLogger)
	assert.NotNil(t, DebugLogger)
}