package checker

import (
	"crypto/md5"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Checker types which implement different types of checks.
type Checker interface {
	// Gather info about the relevant state of the system.
	Collect(config map[string]string)
	// Compare two string lines of system state.
	Compare(a, b string) (int, error)
	Progress() float64
	GetCollected() ([]string, error)
}

// FileChecker collects files under a given path. It can get their size and/or
// calculate md5 hash for each of them. Also paths that should be excluded from
// results can be specified in the config map for Collect function.
type FileChecker struct {
	// Tracks progress of the collection operation, since some can take a while.
	progress float64
	// Collected state of the system
	collected []string
}

// Collects state of all files and dirs under a given path.
// Configuration expects path as the target under which to search, list
// of directories/or files to skip (column separated string), size flag
// to collect file sizes, and hash flag to calculate file hashes.
// Returns list of strings containg dir/file full name, and optionally size or
// hash, comma separated.
func (fc *FileChecker) Collect(config map[string]string) {
	skips := strings.Split(config["skips"], ":")
	collectSize := config["size"] == "true" || config["size"] == "yes"
	collectHash := config["hash"] == "true" || config["hash"] == "yes"
	targetPath := config["path"]
	// TODO create skip map
}

type HashRes struct {
	FileName string
	Hash     string
	Size     int64
}

type FileCheckerPath struct {
	Path        string
	Skips       map[string]bool
	SkipHashing bool //just size checking
}

// List directories and files with their hashes.
func (fc FileChecker) List(root string, skips map[string]bool) (hr []HashRes, err error) {
	// TODO mora postojati skip dir lista
	err = filepath.Walk(root, func(path string, info os.FileInfo, err0 error) error {
		if info.IsDir() {
			if skips[path] {
				return filepath.SkipDir
			}
			hr = append(hr, HashRes{FileName: path})
		} else {
			f, err2 := os.Open(path)
			if err2 != nil {
				return err2
			}
			defer f.Close()
			h := md5.New()
			if _, err3 := io.Copy(h, f); err3 != nil {
				log.Fatal(err3)
			}
			hr = append(hr, HashRes{FileName: path, Hash: string(h.Sum(nil))})
		}
		return nil
	})
	return hr, err
}

type ACLChecker struct {
	progress  float64
	collected []string
}
