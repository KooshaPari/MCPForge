package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Save original writer to restore after test
	originalWriter := Writer
	originalLevels := make(map[Component]LogLevel)
	maps.Copy(originalLevels, ComponentLevels)

	// Set up a buffer to capture logs
	var buf bytes.Buffer
	SetWriter(&buf)

	// Reset buffer and log levels after test
	defer func() {
		SetWriter(originalWriter)
		maps.Copy(ComponentLevels, originalLevels)
	}()

	// Test different log levels
	tests := []struct {
		name           string
		component      Component
		componentLevel LogLevel
		logFunc        func(Logger)
		level          LogLevel
		shouldLog      bool
	}{
		{
			name:           "Debug message with Debug level",
			component:      Core,
			componentLevel: LevelDebug,
			logFunc:        func(l Logger) { l.Debug("test debug message") },
			level:          LevelDebug,
			shouldLog:      true,
		},
		{
			name:           "Debug message with Info level",
			component:      Core,
			componentLevel: LevelInfo,
			logFunc:        func(l Logger) { l.Debug("test debug message") },
			level:          LevelDebug,
			shouldLog:      false,
		},
		{
			name:           "Info message with Info level",
			component:      LSP,
			componentLevel: LevelInfo,
			logFunc:        func(l Logger) { l.Info("test info message") },
			level:          LevelInfo,
			shouldLog:      true,
		},
		{
			name:           "Warn message with Error level",
			component:      Watcher,
			componentLevel: LevelError,
			logFunc:        func(l Logger) { l.Warn("test warn message") },
			level:          LevelWarn,
			shouldLog:      false,
		},
		{
			name:           "Error message with Error level",
			component:      Tools,
			componentLevel: LevelError,
			logFunc:        func(l Logger) { l.Error("test error message") },
			level:          LevelError,
			shouldLog:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Set component log level
			SetLevel(tt.component, tt.componentLevel)

			// Create logger and log message
			logger := NewLogger(tt.component)
			tt.logFunc(logger)

			// Check if message was logged
			loggedMessage := buf.String()
			if tt.shouldLog && loggedMessage == "" {
				t.Errorf("Expected log message but got none")
			} else if !tt.shouldLog && loggedMessage != "" {
				t.Errorf("Expected no log message but got: %s", loggedMessage)
			}

			// When log should appear, check if it contains expected parts
			if tt.shouldLog {
				if !strings.Contains(loggedMessage, tt.level.String()) {
					t.Errorf("Log message missing level '%s': %s", tt.level, loggedMessage)
				}
				if !strings.Contains(loggedMessage, string(tt.component)) {
					t.Errorf("Log message missing component '%s': %s", tt.component, loggedMessage)
				}
			}
		})
	}
}

// snapshotLoggerState captures the package-level logging configuration so a
// test can restore it on cleanup, regardless of how many setup helpers it
// calls or in what order.
type snapshotLoggerState struct {
	jsonOutput     bool
	writer         io.Writer
	testOutput     io.Writer
	defaultLevel   LogLevel
	componentLevels map[Component]LogLevel
}

func snapshotState() snapshotLoggerState {
	logMu.Lock()
	defer logMu.Unlock()
	cl := make(map[Component]LogLevel, len(ComponentLevels))
	for k, v := range ComponentLevels {
		cl[k] = v
	}
	return snapshotLoggerState{
		jsonOutput:      JSONOutput,
		writer:          Writer,
		testOutput:      TestOutput,
		defaultLevel:    DefaultMinLevel,
		componentLevels: cl,
	}
}

func (s snapshotLoggerState) restore() {
	logMu.Lock()
	defer logMu.Unlock()
	JSONOutput = s.jsonOutput
	Writer = s.writer
	TestOutput = s.testOutput
	DefaultMinLevel = s.defaultLevel
	ComponentLevels = make(map[Component]LogLevel, len(s.componentLevels))
	maps.Copy(ComponentLevels, s.componentLevels)
}

// TestLoggerSetupIdempotency verifies that the configuration helpers can be
// invoked repeatedly without changing the observable state of the logger.
// Each helper is called twice (or more) with the same arguments and the
// resulting state is compared to the state after a single call.
func TestLoggerSetupIdempotency(t *testing.T) {
	snap := snapshotState()
	t.Cleanup(snap.restore)

	var buf bytes.Buffer

	// SetJSONOutput is idempotent.
	SetJSONOutput(true)
	stateAfterFirst := JSONOutput
	SetJSONOutput(true)
	if JSONOutput != stateAfterFirst {
		t.Fatalf("SetJSONOutput(true) called twice: state changed from %v to %v", stateAfterFirst, JSONOutput)
	}
	if !JSONOutput {
		t.Fatalf("expected JSONOutput=true after SetJSONOutput(true), got false")
	}

	SetJSONOutput(false)
	SetJSONOutput(false)
	if JSONOutput {
		t.Fatalf("expected JSONOutput=false after SetJSONOutput(false), got true")
	}

	// SetLevel is idempotent for the same (component, level) pair.
	SetLevel(LSP, LevelWarn)
	firstLSP := ComponentLevels[LSP]
	SetLevel(LSP, LevelWarn)
	if ComponentLevels[LSP] != firstLSP {
		t.Fatalf("SetLevel(LSP, LevelWarn) twice: level changed from %v to %v", firstLSP, ComponentLevels[LSP])
	}
	if ComponentLevels[LSP] != LevelWarn {
		t.Fatalf("ComponentLevels[LSP] = %v, want %v", ComponentLevels[LSP], LevelWarn)
	}

	// SetGlobalLevel is idempotent.
	SetGlobalLevel(LevelDebug)
	firstDefault := DefaultMinLevel
	SetGlobalLevel(LevelDebug)
	if DefaultMinLevel != firstDefault {
		t.Fatalf("SetGlobalLevel(LevelDebug) twice: default level changed from %v to %v", firstDefault, DefaultMinLevel)
	}
	if DefaultMinLevel != LevelDebug {
		t.Fatalf("DefaultMinLevel = %v, want %v", DefaultMinLevel, LevelDebug)
	}

	// SetWriter is idempotent for the same writer.
	SetWriter(&buf)
	SetWriter(&buf)
	if Writer != io.Writer(&buf) {
		t.Fatalf("Writer not stable after two SetWriter calls with the same writer")
	}

	// SetupTestLogging and ResetTestLogging can be called repeatedly.
	var capture bytes.Buffer
	SetupTestLogging(&capture)
	SetupTestLogging(&capture)
	if TestOutput == nil {
		t.Fatalf("TestOutput should remain set after SetupTestLogging called twice")
	}
	ResetTestLogging()
	ResetTestLogging()
	if TestOutput != nil {
		t.Fatalf("TestOutput should be nil after ResetTestLogging called twice")
	}
}

// TestLoggerJSONOutputFormat verifies that when JSON output is enabled, each
// emitted log line is a single, well-formed JSON object containing the
// expected structured fields (time, level, component, message).
func TestLoggerJSONOutputFormat(t *testing.T) {
	snap := snapshotState()
	t.Cleanup(snap.restore)

	var capture bytes.Buffer
	SetJSONOutput(true)
	SetupTestLogging(&capture)
	SetLevel(Core, LevelInfo)

	logger := NewLogger(Core)
	logger.Info("hello %s", "world")

	raw := capture.String()
	if raw == "" {
		t.Fatalf("expected JSON log output, got empty buffer")
	}

	// A single record should be a single line.
	line := strings.TrimRight(raw, "\n")
	if line == "" {
		t.Fatalf("expected non-empty JSON line, got whitespace only")
	}
	if strings.Contains(line, "\n") {
		t.Fatalf("expected a single JSON line, got multi-line output: %q", line)
	}

	// The line must be valid JSON.
	var entry map[string]any
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, line)
	}

	// Required structured fields.
	required := []string{"time", "level", "component", "message"}
	for _, key := range required {
		if _, ok := entry[key]; !ok {
			t.Errorf("JSON output missing required field %q: %s", key, line)
		}
	}

	if got, _ := entry["level"].(string); got != LevelInfo.String() {
		t.Errorf("level = %q, want %q", got, LevelInfo.String())
	}
	if got, _ := entry["component"].(string); got != string(Core) {
		t.Errorf("component = %q, want %q", got, Core)
	}
	if got, _ := entry["message"].(string); got != "hello world" {
		t.Errorf("message = %q, want %q", got, "hello world")
	}

	// The time field, if present, must be a string (RFC3339Nano).
	if ts, ok := entry["time"].(string); !ok || ts == "" {
		t.Errorf("time field should be a non-empty string, got %v (type %T)", entry["time"], entry["time"])
	}
}
