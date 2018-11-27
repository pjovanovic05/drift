package checker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
	"sync"
)

type UserChecker struct {
	BasicChecker
	mu sync.Mutex
}

// Collect user info from /etc/passwd files.
func (uc *UserChecker) Collect(config map[string]string) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		log.Fatal(err)
	}
	defer passwd.Close()
	var users []string
	lines, err := ioutil.ReadAll(passwd)
	if err != nil {
		uc.err = err
		return
	}
	for _, line := range lines {
		if comment := strings.Index(string(line), "#"); comment >= 0 {
			continue
		}
		comps := strings.Split(string(line), ":")
		if len(comps) > 0 {
			users = append(users, comps[0])
		}
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	for _, name := range users {
		usr, err := user.Lookup(name)
		if err != nil {
			log.Fatal(err)
		}
		valueline := fmt.Sprintf("%s, %s, %s", usr.Uid, usr.Gid, usr.HomeDir)
		groups, err := usr.GroupIds()
		if err != nil {
			log.Fatal(err)
		}
		gs := strings.Join(groups, ",")
		valueline = valueline + "," + gs
		uc.collected = append(uc.collected, Pair{Key: usr.Username, Value: valueline})
	}
	// TODO: sort collected
	uc.progress = "done"
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
