package checker

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Pair is a key-value pair to hold intermediate state collection results.
type Pair struct {
	Key   string
	Value string
}

func (p Pair) String() string {
	return p.Key + ":" + p.Value
}

// Checker types which implement different types of checks.
type Checker interface {
	// Gather info about the relevant state of the system. (async)
	Collect(config map[string]string)
	// Get current progress of the collection.
	Progress() string
	// Fetch the collected system state info.
	GetCollected() ([]Pair, error)
	// Get error if any has occured during collection.
	GetErr() error
}

type BasicChecker struct {
	// Tracks progress of the collection operation, since some can take a while.
	progress string
	// Collected state of the system
	collected []Pair
	err       error
}

// FileChecker collects files under a given path. It can get their size and/or
// calculate md5 hash for each of them. Also paths that should be excluded from
// results can be specified in the config map for Collect function.
type FileChecker struct {
	BasicChecker
	mu sync.Mutex //protects progress and collected
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
			fc.mu.Lock()
			fc.collected = append(fc.collected, Pair{Key: path, Value: "DIR"})
			fc.progress = "checking: " + path
			fc.mu.Unlock()
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
				recline = fmt.Sprintf("%d,%s", info.Size(), string(h.Sum(nil)))
			} else {
				recline = fmt.Sprintf("%d", info.Size())
			}
			fc.mu.Lock()
			fc.collected = append(fc.collected, Pair{Key: path, Value: recline})
			fc.mu.Unlock()
		}
		return nil
	})
	fc.mu.Lock()
	fc.progress = "sorting..."
	fc.mu.Unlock()
	sort.SliceStable(fc.collected, func(i, j int) bool {
		return fc.collected[i].Key < fc.collected[j].Value
	})
	fc.mu.Lock()
	fc.progress = "done"
	fc.mu.Unlock()
}

func (fc *FileChecker) Progress() string {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.progress
}

func (fc *FileChecker) GetError() error {
	return fc.err
}

// GetCollected returns array of pairs with collected state.
func (fc *FileChecker) GetCollected() ([]Pair, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.collected, nil
}

type UserChecker struct {
	BasicChecker
}

type ACLChecker struct {
	BasicChecker
}

type PackageChecker struct {
	BasicChecker
}

func (pc *PackageChecker) Collect(config map[string]string) {
	// TODO: rpm/yum/dnf or apt or zypper or pacman or whatever?
}
