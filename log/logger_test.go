package log

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-errors/errors"
)

func TestLogger_Errorf(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{fileWriter: &buf}
		testErr := errors.New("yet another test error")
		logger.Errorf("Error details: %v, some id: %d", testErr, 456)

		output := buf.String()
		if !strings.Contains(output, "=== Stack Trace Start ===") {
			t.Errorf("expected stack trace to be included error is in argument list, but it was not:\n%s", output)
		}
		if !strings.Contains(output, "yet another test error") {
			t.Errorf("expected error message to be included in the stack trace, but it was not:\n%s", output)
		}
		stackTraceRegex := regexp.MustCompile(`\n\t.*`)
		if !stackTraceRegex.MatchString(output) {
			t.Errorf("expected detailed stack trace information when error is in argument list, but it seems to be missing:\n%s", output)
		}
	})

	t.Run("wrapped error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{fileWriter: &buf}
		testErr := fmt.Errorf("wrapped %w error", errors.New("yet another test error"))
		logger.Errorf("Error details: %v, processing id: %d", testErr, 456)

		output := buf.String()
		if !strings.Contains(output, "=== Stack Trace Start ===") {
			t.Errorf("expected stack trace to be included when error is wrapped, but it was not:\n%s", output)
		}
		stackTraceRegex := regexp.MustCompile(`\n\t.*`)
		if !stackTraceRegex.MatchString(output) {
			t.Errorf("expected detailed stack trace information when error is wrapped, but it seems to be missing:\n%s", output)
		}
	})

	t.Run("without error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{fileWriter: &buf}
		logger.Errorf("Hello %s", "World")

		output := buf.String()
		if strings.Contains(output, "=== Stack Trace Start ===") {
			t.Errorf("expected stack trace to NOT be included, but it was:\n%s", output)
		}
		if strings.Contains(output, "=== Stack Trace End ===") {
			t.Errorf("expected stack trace end marker to NOT be included, but it was:\n%s", output)
		}
	})
}
