package logger

import (
	"bytes"
	"log"
	"testing"
)

func TestStdLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "", 0),
	}
	l.Info("test message", "key", "value")
	output := buf.String()
	expected := "[INFO] test message [key value]\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestStdLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "", 0),
	}
	l.Warn("test message", "key", "value")
	output := buf.String()
	expected := "[WARN] test message [key value]\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestStdLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "", 0),
	}
	l.Error("test message", "key", "value")
	output := buf.String()
	expected := "[ERROR] test message [key value]\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestNewStdLogger(t *testing.T) {
	l := NewStdLogger()
	if l == nil {
		t.Error("Expected NewStdLogger to return a non-nil logger")
	}
	if l.logger == nil {
		t.Error("Expected StdLogger.logger to be non-nil")
	}
}

func TestNewTestLogger(t *testing.T) {
	l := NewTestLogger()
	if l == nil {
		t.Error("Expected NewTestLogger to return a non-nil logger")
	}
	if l.logger == nil {
		t.Error("Expected StdLogger.logger to be non-nil")
	}
}
