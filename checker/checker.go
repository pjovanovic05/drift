package checker

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Checker types which implement different types of checks.
type Checker interface {
	// Gather info about the relevant state of the system. (async)
	Collect(config map[string]string)
	// Compare two string lines of system state.
	Compare(a, b string) (int, error)
	// Get current progress of the collection.
	Progress() float64
	// Fetch the collected system state info.
	GetCollected() ([]string, error)
	// Get error if any has occured during collection.
	GetErr() error
}

type BasicChecker struct {
	// Tracks progress of the collection operation, since some can take a while.
	progress float64
	// Collected state of the system
	collected []string
	err       error
}

// FileChecker collects files under a given path. It can get their size and/or
// calculate md5 hash for each of them. Also paths that should be excluded from
// results can be specified in the config map for Collect function.
type FileChecker struct {
	BasicChecker
}

// Collect state of all files and dirs under a given path.
// Configuration expects path as the target under which to search, list
// of directories/or files to skip (column separated string), and
// hash flag to calculate file hashes.
// Returns list of strings containg dir/file full name, size or
// hash, comma separated.
func (fc *FileChecker) Collect(config map[string]string) {
	skipPaths := strings.Split(config["skips"], ":")
	collectHash := config["hash"] == "true" || config["hash"] == "yes"
	targetPath := config["path"]
	// create skip map
	skips := make(map[string]bool)
	for _, dir := range skipPaths {
		skips[dir] = true
	}
	fc.err = filepath.Walk(targetPath, func(path string, info os.FileInfo, err0 error) error {
		if info.IsDir() {
			if skips[path] {
				return filepath.SkipDir
			}
			fc.collected = append(fc.collected, path)
			fc.progress += 1.0
		} else {
			var recline string
			if skips[path] {
				return nil
			}
			if collectHash {
				f, err2 := os.Open(path)
				if err2 != nil {
					// TODO sta sa greskom?
					return err2
				}
				defer f.Close()
				h := md5.New()
				if _, err3 := io.Copy(h, f); err3 != nil {
					log.Fatal(err3) // TODO jel ovo ok?
				}
				recline = fmt.Sprintf("%s,%d,%s", path, info.Size(), string(h.Sum(nil)))
			} else {
				recline = fmt.Sprintf("%s,%d", path, info.Size())
			}
			fc.collected = append(fc.collected, recline)
		}
		return nil
	})
}

func (fc *FileChecker) Progress() float64 {
	return 0.0
}

func (fc *FileChecker) Compare(a, b string) (int, error) {

}

func (fc *FileChecker) GetError() error {
	return fc.err
}

type ACLChecker struct {
	BasicChecker
}
