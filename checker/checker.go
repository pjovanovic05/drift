package checker

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	fc.mu.Lock()
	fc.collected = fc.collected[:0]
	fc.mu.Unlock()
	skipPaths := strings.Split(config["skips"], ":")
	collectHash := config["hash"] == "true" || config["hash"] == "yes"
	targetPath := config["path"]
	// create skip map
	skips := make(map[string]bool)
	for _, dir := range skipPaths {
		skips[dir] = true
	}
	fc.err = filepath.Walk(targetPath, func(path string, info os.FileInfo, err0 error) error {
		var recline string
		if skips[path] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			recline = "DIR"
		} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			recline = "SYMLINK"
		} else {
			if collectHash && isFileReadable(&info) {
				f, err2 := os.OpenFile(path, os.O_RDONLY, 0666)
				if err2 != nil {
					return err2
				}
				defer f.Close()
				h := md5.New()
				if _, err3 := io.Copy(h, f); err3 != nil {
					log.Fatal(err3) // TODO jel ovo ok?
				}
				recline = fmt.Sprintf("%d,%x", info.Size(), h.Sum(nil))
			} else {
				recline = fmt.Sprintf("%d", info.Size())
			}
		}
		fc.mu.Lock()
		fc.collected = append(fc.collected, Pair{Key: path, Value: recline})
		fc.progress = "checked: " + path
		fc.mu.Unlock()
		return nil
	})

	fc.mu.Lock()
	fc.progress = "sorting..."
	fc.mu.Unlock()
	sort.SliceStable(fc.collected, func(i, j int) bool {
		return fc.collected[i].Key < fc.collected[j].Key
	})
	fc.mu.Lock()
	fc.progress = "done"
	fc.mu.Unlock()
}

func isFileReadable(info *os.FileInfo) bool {
	return (*info).Mode().String()[1] == 'r'
}

func (fc *FileChecker) Progress() string {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.progress
}

func (fc *FileChecker) GetErr() error {
	return fc.err
}

// GetCollected returns array of pairs with collected state.
func (fc *FileChecker) GetCollected() ([]Pair, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.collected, fc.err
}

type UserChecker struct {
	BasicChecker
}

type ACLChecker struct {
	BasicChecker
}

// PackageChecker collects the list of packages and versions.
type PackageChecker struct {
	BasicChecker
}

// Collect installed package names and versions.
func (pmc *PackageChecker) Collect(config map[string]string) {
	if strings.ToLower(config["manager"]) != "rpm" {
		pmc.err = errors.New("Only RPM manager is supported.")
		// TODO support for other package managers.
		return
	}
	// rpm -qa --queryformat "%{NAME},%{VERSION}\n"
	output, err := exec.Command("rpm", "-qa", "--queryformat",
		`"%{NAME},%{VERSION}\n"`).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(output), "\n")
	// convert csv output string to list of packages => collected
	for _, line := range lines {
		kv := strings.Split(line, ",")
		pmc.collected = append(pmc.collected, Pair{Key: kv[0], Value: kv[1]})
	}
	// sort collected
	sort.SliceStable(pmc.collected, func(i, j int) bool {
		return pmc.collected[i].Key < pmc.collected[j].Key
	})
	pmc.progress = "done"
}

func (pmc *PackageChecker) Progress() string {
	return pmc.progress
}

func (pmc *PackageChecker) GetCollected() ([]Pair, error) {
	return pmc.collected, pmc.err
}

func (pmc *PackageChecker) GetErr() error {
	return pmc.err
}
