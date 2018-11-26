package checker

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ACLChecker struct {
	BasicChecker
	mu sync.Mutex
}

func (aclc *ACLChecker) Collect(config map[string]string) {
	// TODO:
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
		aclc.mu.Lock()
		B
		aclc.collected = append(aclc.collected, Pair{Key: path, Value: info.Mode().String()})
		aclc.progress = path
		aclc.mu.Unlock()
		// TODO: get owner uid and gid:
		// https://groups.google.com/forum/#!topic/golang-nuts/ywS7xQYJkHY
	})
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
