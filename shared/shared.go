package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// Abs takes a path and, if it is not already absolute, makes it absolute with
// the assumption that it is relative to the root directory of the project.
func Abs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Sprint("shared.Abs: could not recover file path to core/abs.go"))
	}

	return filepath.Join(filepath.Dir(file), "..", path)
}

// Time returns the current time in UTC rounded to the nearest millisecond.
func Time() time.Time {
	return time.Now().Round(time.Millisecond).UTC()
}

// GetFileInfo returns the file information for a path and panics if any errors
// occur.
func GetFileInfo(path string) os.FileInfo {
	if file, err := os.Open(Abs(path)); err != nil {
		log.Panicf("GetFileInfo: got error while opening %s: %s", path, err)
	} else if info, err := file.Stat(); err != nil {
		log.Panicf("GetFileInfo: got error while fetching file information for %s: %s", path, err)
	} else {
		return info
	}
	return nil
}
