package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
)

// ACLChecker collects info about file access rights.
type ACLChecker struct {
	BasicChecker
	mu sync.Mutex
}

// Collect gathers info about file access rights. Takess
// configuration params:
// psth - target path
// skips = column (:) separated list of paths to skip
// Returns:
// key: path, value: acl, uid, gid
// Remark: works only on linux
func (aclc *ACLChecker) Collect(config map[string]string) {
	aclc.mu.Lock()
	aclc.collected = aclc.collected[:0]
	aclc.mu.Unlock()
	skipPaths := strings.Split(config["skips"], ":")
	targetPath := config["path"]

	// create skip map
	skips := make(map[string]bool)
	for _, dir := range skipPaths {
		skips[dir] = true
	}

	aclc.err = filepath.Walk(targetPath, func(path string, info os.FileInfo, err0 error) error {
		if skips[path] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		uid := info.Sys().(*syscall.Stat_t).Uid
		gid := info.Sys().(*syscall.Stat_t).Gid
		recline := fmt.Sprintf("%s, %d, %d", info.Mode().String(), uid, gid)
		aclc.mu.Lock()
		aclc.collected = append(aclc.collected, Pair{Key: path, Value: recline})
		aclc.progress = path
		aclc.mu.Unlock()
		return nil
	})
	aclc.mu.Lock()
	aclc.progress = "sorting..."
	aclc.mu.Unlock()
	sort.SliceStable(aclc.collected, func(i, j int) bool {
		return aclc.collected[i].Key < aclc.collected[j].Key
	})
	aclc.mu.Lock()
	aclc.progress = "acl collection done"
	aclc.mu.Unlock()
}

func (aclc *ACLChecker) Progress() string {
	aclc.mu.Lock()
	defer aclc.mu.Unlock()
	return aclc.progress
}

func (aclc *ACLChecker) GetCollected() ([]Pair, error) {
	aclc.mu.Lock()
	defer aclc.mu.Unlock()
	return aclc.collected, aclc.err
}

func (aclc *ACLChecker) GetErr() error {
	aclc.mu.Lock()
	defer aclc.mu.Unlock()
	return aclc.err
}
