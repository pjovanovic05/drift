package checker

import (
	"errors"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

// PackageChecker collects the list of packages and versions.
type PackageChecker struct {
	BasicChecker
	mu sync.Mutex
}

// Collect installed package names and versions.
func (pmc *PackageChecker) Collect(config map[string]string) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()
	if strings.ToLower(config["manager"]) != "rpm" {
		pmc.err = errors.New("Only RPM manager is supported.")
		// TODO support for other package managers.
		return
	}
	// rpm -qa --queryformat "%{NAME},%{VERSION}\n"
	output, err := exec.Command("rpm", "-qa", "--queryformat",
		`%{NAME},%{VERSION}\n`).CombinedOutput()
	if err != nil {
		pmc.err = err
		// log.Fatal(err)
	}
	lines := strings.Split(string(output), "\n")
	// convert csv output string to list of packages => collected
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// log.Println("collected:", line)
		kv := strings.Split(line, ",")
		pmc.collected = append(pmc.collected, Pair{Key: kv[0], Value: kv[1]})
	}
	// sort collected
	sort.SliceStable(pmc.collected, func(i, j int) bool {
		return pmc.collected[i].Key < pmc.collected[j].Key
	})
	pmc.progress = "package collection done"
}

func (pmc *PackageChecker) Progress() string {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()
	return pmc.progress
}

func (pmc *PackageChecker) GetCollected() ([]Pair, error) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()
	return pmc.collected, pmc.err
}

func (pmc *PackageChecker) GetErr() error {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()
	return pmc.err
}
