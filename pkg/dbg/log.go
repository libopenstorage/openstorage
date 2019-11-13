package dbg

import (
	"os"

	"github.com/sirupsen/logrus"
)

func LogErrorAndPanicf(err error, format string, args ...interface{}) {
	logrus.Errorf("LogErrorAndPanicf: error: %v", err)
	panicf(format, args...)

}

// Panicf outputs error message, dumps threads and exits.
func Panicf(format string, args ...interface{}) {
	panicf(format, args...)
}

// Panicf outputs error message, dumps threads and exits.
func panicf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
	err := DumpGoProfile()
	if err != nil {
		logrus.Fatal(err)
	}
	DumpHeap()
	os.Exit(6)
}

// Assert Panicf's if the condition evaluates to false.
func Assert(condition bool, format string, args ...interface{}) {
	if !condition {
		Panicf(format, args...)
	}
}
