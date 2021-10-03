package session

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/knchan0x/umeshu/log"
)

// inMemory is the default implementation of Store Interface.
// It is thread-safe.
type inMemory struct {
	cache map[string]*list.Element
	quene *list.List // oldest in the front
	mu    sync.RWMutex
}

// init registers constructor in storeConstructorMap.
func init() {
	Register("InMemory", newInMemoryStore)
}

// newLRUStore returns a store object.
func newInMemoryStore() Store {
	return &inMemory{
		quene: list.New(),
	}
}

// Read returns session object by session id, return nil
// if no such session id
func (m *inMemory) Read(sid string) (Session, error) {
	m.mu.RLock()
	defer m.mu.Unlock()

	if element, ok := m.cache[sid]; ok {
		if err := element.Value.(Session).Set("LastAccessTime", time.Now()); err == nil {
			return element.Value.(Session), nil
		}
	}
	return nil, fmt.Errorf("session id not exists.")
}

// Insert creates new session object according to session id and token
// and insert it into cache. token is the type of User-Agent.
func (m *inMemory) Insert(sid string, token string) (Session, error) {
	if m.quene == nil {
		m.quene = list.New()
	}

	// config session
	newSession := make(session)
	if err := newSession.Set("SID", sid); err != nil {
		return nil, err
	}
	if err := newSession.Set("Token", token); err != nil {
		return nil, err
	}
	if err := newSession.Set("CreateTime", time.Now()); err != nil {
		return nil, err
	}
	if err := newSession.Set("LastAccessTime", time.Now()); err != nil {
		return nil, err
	}

	m.mu.Lock()
	// add to cache
	element := m.quene.PushBack(newSession)
	m.cache[sid] = element
	m.mu.Unlock()

	return newSession, nil
}

// UpdateSID replaces old session id by new id.
func (m *inMemory) UpdateSID(old string, new string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	element, ok := m.cache[old]
	if !ok {
		return
	}
	element.Value.(Session).Set("SID", new)
	m.cache[new] = element
	delete(m.cache, old)

	return
}

// Delete deletes session according to session id.
func (m *inMemory) Delete(sid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.quene == nil {
		panic("session persistence: cache not exists.")
	}

	element, ok := m.cache[sid]
	if !ok {
		return fmt.Errorf("key not exists.")
	}
	delete(m.cache, sid)
	m.quene.Remove(element)
	return nil
}

// GC forces to remove session objects according to lifetime.
func (m *inMemory) GC(maxLifeTime int) {
	if m.quene == nil {
		return
	}

	for {
		element := m.quene.Front()
		// no element in cache
		if element == nil || element.Value == nil {
			break
		}
		// no last access time, unable to clean cache according to life time
		lastAccessTime := element.Value.(Session).Get("LastAccessTime")
		if lastAccessTime == nil {
			break
		}
		// life time is shorter than maxmium, end GC
		if (lastAccessTime.(time.Time).Unix() + int64(maxLifeTime)) > time.Now().Unix() {
			break
		}

		s := element.Value.(Session).Get("SID")
		if s == nil {
			continue
		}
		sid := s.(string)
		if err := m.Delete(sid); err == nil {
			log.Error("unable to delete session ID: %s", sid)
		}
	}
}
