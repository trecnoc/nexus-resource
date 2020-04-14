package utils

import (
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
	"github.com/sirupsen/logrus"
)

// Fatal is used to print an error message and exit the application.
func Fatal(doing string, err error) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	os.Exit(1)
}

// Sayf is used to print a message on Stderr
func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

// StandardLogger enforces log message formats
type StandardLogger struct {
	logger  *logrus.Logger
	enabled bool
}

// NewLogger initializes the standard logger
func NewLogger(enabled bool) *StandardLogger {
	var baseLogger = logrus.New()
	var standardLogger = &StandardLogger{
		logger:  baseLogger,
		enabled: true,
	}
	standardLogger.logger.Formatter = &logrus.TextFormatter{}
	logfile, _ := os.OpenFile("/tmp/concourse-nexus-resource.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	standardLogger.logger.Out = logfile
	return standardLogger
}

// NewNexusClient logs when a new Nexus is created
func (l *StandardLogger) NewNexusClient(nexusURL string, username string) {
	if l.enabled {
		l.logger.WithFields(logrus.Fields{
			"url":  nexusURL,
			"user": username,
		}).Info("New Nexus Client created")
	}
}

// LogSimpleMessage according to a format specifier
func (l *StandardLogger) LogSimpleMessage(message string, a ...interface{}) {
	if l.enabled {
		l.logger.Info(fmt.Sprintf(message, a...))
	}
}

// LogHTTPRequest logs and Http Request
func (l *StandardLogger) LogHTTPRequest(requestType string, requestURL string) {
	if l.enabled {
		l.logger.WithFields(logrus.Fields{
			"type": requestType,
			"url":  requestURL,
		}).Info("Executing Http Request")
	}
}
