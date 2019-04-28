package core

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

// Abs takes a path and, if it is not already absolute, makes it absolute with
// the assumption that it is relative to the root directory of the project.
func Abs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Sprint("core.Abs: could not recover file path to core/abs.go"))
	}

	return filepath.Join(filepath.Dir(file), "..", path)
}

// Time returns the current time in UTC rounded to the nearest millisecond.
func Time() time.Time {
	return time.Now().Round(time.Millisecond).UTC()
}
