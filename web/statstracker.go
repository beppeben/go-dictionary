package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/roylee0704/gron"
	"github.com/roylee0704/gron/xtime"
)

type User struct {
	Ip            string
	City          string
	Region        string
	Country       string
	SearchedWords map[string]bool
	Counter       int64
	Referer       string
	UserAgent     string
}

type StatsTracker struct {
	mutex     sync.Mutex
	keyToUser map[string]*User
	eutils    EmailUtils
}

func NewStatsTracker(e EmailUtils) *StatsTracker {
	stats := &StatsTracker{eutils: e}
	stats.keyToUser = make(map[string]*User)

	c := gron.New()
	c.AddFunc(gron.Every(1*xtime.Day).At("16:00"), func() {
		stats.sendSummaryAndClear()
	})
	c.Start()

	return stats
}

func (stats *StatsTracker) sendSummaryAndClear() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	var buffer bytes.Buffer
	buffer.WriteString("Users: " + strconv.Itoa(len(stats.keyToUser)) + ".")
	for _, user := range stats.keyToUser {
		buffer.WriteString("\n\n")
		buffer.WriteString(user.Ip + " (" + user.City + ", " + user.Country + "). ")
		buffer.WriteString("Agent: " + user.UserAgent + ". ")
		buffer.WriteString("Referer: " + user.Referer + ". ")
		buffer.WriteString("Hits: " + strconv.FormatInt(user.Counter, 10) + ". ")
		if len(user.SearchedWords) > 0 {
			buffer.WriteString("Words:")
			i := 0
			for word, _ := range user.SearchedWords {
				buffer.WriteString(" " + word)
				if i < len(user.SearchedWords)-1 {
					buffer.WriteString(",")
				}
			}
			buffer.WriteString(".")
		}
	}
	stats.eutils.SendEmailToAdmins("Daily Report", buffer.String())
	stats.keyToUser = make(map[string]*User)
}

func (stats *StatsTracker) getOrAddUser(usr *User, key string) *User {
	user := stats.keyToUser[key]
	if user == nil {
		user = usr
		user.SearchedWords = make(map[string]bool)
		stats.keyToUser[key] = user
	}
	user.Counter++
	if user.City == "" {
		user.getLocation()
	}
	return user
}

/*
func (stats *StatsTracker) getOrAddUser(ip string, key string) *User {
	user := stats.keyToUser[key]
	if user == nil {
		user = &User{Ip: ip}
		user.SearchedWords = make(map[string]bool)
		stats.keyToUser[key] = user
	}
	user.Counter++
	if user.City == "" {
		user.getLocation()
	}
	return user
}
*/
/*
func (stats *StatsTracker) NotifyUser(ip string, key string, word string) {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	user := stats.getOrAddUser(ip, key)
	if word != "" {
		user.SearchedWords[word] = true
	}
}
*/

func (stats *StatsTracker) NotifyUser(usr *User, key string, word string) {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	user := stats.getOrAddUser(usr, key)
	if word != "" {
		user.SearchedWords[word] = true
	}
}

func (user *User) getLocation() error {
	r, err := http.Get("http://ipinfo.io/" + user.Ip + "/json")
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		return err
	}

	return nil
}
