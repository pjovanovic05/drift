package checker

import "sync"

type ACLChecker struct {
	BasicChecker
	mu sync.Mutex
}

func (aclc *ACLChecker) Collect(config map[string]string) {
	// TODO:
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
