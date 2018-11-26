package checker

import (
	"os"
		"log"
	"sync"
)

type UserChecker struct {
	BasicChecker
	mu sync.Mutex
}

//Collect user and group info from /etc passwd and group files.
func (uc *UserChecker) Collect(config map[string]string) {
	// TODO: parse users and groups from /etc/passwd and /etc/group
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		log.Fatal(err)
	}
	defer passwd.Close()
}

func (uc *UserChecker) Progress() string {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.progress
}

func (uc *UserChecker) GetCollected() ([]Pair, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.collected, uc.err
}

func (uc *UserChecker) GetErr() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.err
}
