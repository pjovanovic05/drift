package checker

import (
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

// Collect user and group info from /etc passwd and group files.
func (uc *UserChecker) Collect(config map[string]string) {
	// TODO: parse users and groups from /etc/passwd and /etc/group
	// https://www.socketloop.com/tutorials/golang-get-all-local-users-and-print-out-their-home-directory-description-and-group-id
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
		// TODO: sta sve od korisnika da sacuvam kao value?
		log.Println(usr.Gid)
	}
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
