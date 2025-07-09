package logger

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSetupLogger_DefaultLevel(t *testing.T) {
	// Unset LOGGER_LEVEL to test default
	os.Unsetenv("LOGGER_LEVEL")
	SetupLogger()

	if log.GetLevel() != log.DebugLevel {
		t.Errorf("Expected log level to be DebugLevel, got %v", log.GetLevel())
	}
}

func TestSetupLogger_CustomLevel(t *testing.T) {
	os.Setenv("LOGGER_LEVEL", "warn")
	defer os.Unsetenv("LOGGER_LEVEL")

	SetupLogger()

	if log.GetLevel() != log.DebugLevel {
		t.Errorf("Expected log level to be DebugLevel, got %v", log.GetLevel())
	}
}

func TestSetupLogger_InvalidLevel(t *testing.T) {
	// Set an invalid LOGGER_LEVEL
	os.Setenv("LOGGER_LEVEL", "invalid_level")
	defer os.Unsetenv("LOGGER_LEVEL")

	// Save and defer restore of original logrus exit function
	origExitFunc := log.StandardLogger().ExitFunc
	defer func() { log.StandardLogger().ExitFunc = origExitFunc }()

	called := false
	log.StandardLogger().ExitFunc = func(int) { called = true }

	// SetupLogger should call log.Fatalf, which triggers ExitFunc
	SetupLogger()

	if !called {
		t.Errorf("Expected logger to call ExitFunc on invalid LOGGER_LEVEL")
	}
}
