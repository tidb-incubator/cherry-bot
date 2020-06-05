package member

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/pingcap-incubator/cherry-bot/config"
)

const (
	cacheTime = 7 * time.Hour
)

// Member struct
// Member is for specify if a user is member
// this does authority control
type Member struct {
	cache  map[string]*user
	github *github.Client
	member config.Member
}

type user struct {
	login      string
	member     bool
	lastUpdate time.Time
}

// New create member
func New(github *github.Client, member config.Member) *Member {
	return &Member{
		cache:  make(map[string]*user),
		github: github,
		member: member,
	}
}

// IfMember make cache and check if a user is member
func (m *Member) IfMember(login string) bool {
	if isMember, cacheValid := m.cacheIfMember(login); cacheValid {
		return isMember
	}
	for _, org := range m.member.Orgs {
		if ifOrgMember, _, err := m.github.Organizations.IsMember(context.Background(), org, login); err == nil {
			if ifOrgMember {
				return m.cacheMember(login, true)
			}
		} else {
			log.Println(err)
		}
	}
	return m.cacheMember(login, false)
}

func (m *Member) cacheIfMember(login string) (bool, bool) {
	if user, ok := m.cache[login]; ok {
		if time.Now().Sub(user.lastUpdate) < cacheTime {
			return user.member, true
		}
	}
	return false, false
}

func (m *Member) cacheMember(login string, member bool) bool {
	m.cache[login] = &user{
		login:      login,
		member:     member,
		lastUpdate: time.Now(),
	}
	return member
}
