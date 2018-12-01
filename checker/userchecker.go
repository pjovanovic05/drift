package checker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"sort"
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
	blines, err := ioutil.ReadAll(passwd)
	if err != nil {
		uc.err = err
		return
	}
	lines := strings.Split(string(blines), "\n")
	for _, line := range lines {
		if comment := strings.Index(line, "#"); comment >= 0 {
			continue
		}
		comps := strings.Split(line, ":")
		if len(comps) > 0 {
			match, err := regexp.MatchString(config["Pattern"], comps[0])
			if err != nil {
				uc.err = err
				break
			}
			if match {
				log.Println("uname:", comps[0])
				users = append(users, comps[0])
			}
		}
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.collected = uc.collected[:0]
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
	sort.SliceStable(uc.collected, func(i, j int) bool {
		return uc.collected[i].Key < uc.collected[j].Key
	})
	uc.progress = "user collection done"
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
