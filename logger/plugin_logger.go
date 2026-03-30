package logger

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type terraformPluginFormatter struct {
	logrus.Formatter
}

func (f *terraformPluginFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Message = "[TerraformPlugin] " + entry.Message
	return f.Formatter.Format(entry)
}

// TerraformPluginLogger adapts logrus to the hclog.Logger interface used by Terraform plugins.
type TerraformPluginLogger struct {
	logger *logrus.Logger
}

// NewTerraformPluginLogger creates a TerraformPluginLogger using the current application log config.
func NewTerraformPluginLogger() TerraformPluginLogger {
	config := getConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.SetReportCaller(false)
	logger.SetFormatter(&terraformPluginFormatter{Formatter: config.Formatter})

	// Disable terraform provider log if we are not in trace level
	if config.Level == logrus.TraceLevel {
		logger.SetLevel(logrus.TraceLevel)
	}

	return TerraformPluginLogger{logger}
}

// Trace logs a message at TRACE level.
func (t TerraformPluginLogger) Trace(msg string, args ...interface{}) {
	t.logger.Trace(msg, args)
}

// Debug logs a message at DEBUG level (routed to TRACE).
func (t TerraformPluginLogger) Debug(msg string, args ...interface{}) {
	t.Trace(msg, args)
}

// Info logs a message at INFO level (routed to TRACE).
func (t TerraformPluginLogger) Info(msg string, args ...interface{}) {
	t.Trace(msg, args)
}

// Warn logs a message at WARN level (routed to TRACE).
func (t TerraformPluginLogger) Warn(msg string, args ...interface{}) {
	t.Trace(msg, args)
}

func (t TerraformPluginLogger) Error(msg string, args ...interface{}) {
	t.Trace(msg, args)
}

// IsTrace always returns true; all plugin output is routed to TRACE.
func (t TerraformPluginLogger) IsTrace() bool {
	return true
}

// IsDebug always returns false.
func (t TerraformPluginLogger) IsDebug() bool {
	return false
}

// IsInfo always returns false.
func (t TerraformPluginLogger) IsInfo() bool {
	return false
}

// IsWarn always returns false.
func (t TerraformPluginLogger) IsWarn() bool {
	return false
}

// IsError always returns false.
func (t TerraformPluginLogger) IsError() bool {
	return false
}

// With returns the logger unchanged (args are ignored).
func (t TerraformPluginLogger) With(_ ...interface{}) hclog.Logger {
	return t
}

// Named returns the logger unchanged (name is ignored).
func (t TerraformPluginLogger) Named(_ string) hclog.Logger {
	return t
}

// ResetNamed returns the logger unchanged (name is ignored).
func (t TerraformPluginLogger) ResetNamed(_ string) hclog.Logger {
	return t
}

// GetLevel always returns hclog.Trace.
func (t TerraformPluginLogger) GetLevel() hclog.Level {
	return hclog.Trace
}

// SetLevel is a no-op; the level is fixed at TRACE.
func (t TerraformPluginLogger) SetLevel(_ hclog.Level) {}

// StandardLogger returns a standard log.Logger writing to the underlying logrus writer.
func (t TerraformPluginLogger) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	stdLogger := log.New(t.logger.Writer(), "", log.Flags())
	stdLogger.SetOutput(t.logger.Writer())
	return stdLogger
}

// StandardWriter returns the underlying logrus writer.
func (t TerraformPluginLogger) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return t.logger.Writer()
}

// Log logs a message at TRACE level regardless of the provided level.
func (t TerraformPluginLogger) Log(_ hclog.Level, msg string, args ...interface{}) {
	t.logger.Log(logrus.TraceLevel, msg, args)
}

// ImpliedArgs always returns nil.
func (t TerraformPluginLogger) ImpliedArgs() []interface{} {
	return nil
}

// Name returns the fixed logger name "TerraformPlugin".
func (t TerraformPluginLogger) Name() string {
	return "TerraformPlugin"
}
