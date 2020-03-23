package utils

import (
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
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
