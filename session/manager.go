package session

import (
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/knchan0x/umeshu/log"
)

// Session manager manages session and session store.
type SessionManager struct {
	store       Store  // session store
	name        string // name of session cookie
	maxLifeTime int    // max lifetime for session
	lapseTime   int    // lapse time for session id

	// GC
	isGCStarted bool
	gcStop      chan struct{}
}

// Settings of session manager.
type ManagerSettings struct {
	Name        string // name of session cookie
	Store       Store  // session store
	MaxLifeTime int    // max lifetime for session, in second
	LapseTime   int    // lapse time for session id, in second
}

const (
	defaultSessionName string = "SID"
	defaultMaxLifeTime int    = 86400 // one day
	defaultLapseTime   int    = 1800  // half hour
)

// DefaultSessionManager provides default values for session manager.
var DefaultSessionManager = ManagerSettings{
	Name:        defaultSessionName,
	Store:       NewStore("LRU_Cache"),
	MaxLifeTime: defaultMaxLifeTime,
	LapseTime:   defaultLapseTime,
}

// Global session manager
var Manager *SessionManager

// NewManager creates and returns new session manager object.
// It will start running GC at separate goroutine.
// SessionManager can only be created once.
func NewManager(settings ManagerSettings) *SessionManager {
	if Manager != nil {
		return Manager
	}

	if settings.Name == "" {
		settings.Name = DefaultSessionManager.Name
	}
	if settings.Store == nil {
		settings.Store = DefaultSessionManager.Store
	}
	if settings.MaxLifeTime == 0 {
		settings.MaxLifeTime = DefaultSessionManager.MaxLifeTime
	}
	if settings.LapseTime == 0 {
		settings.LapseTime = DefaultSessionManager.MaxLifeTime
	}

	Manager := &SessionManager{
		name:        settings.Name,
		store:       settings.Store,
		maxLifeTime: settings.MaxLifeTime,
		lapseTime:   settings.LapseTime,
	}

	Manager.StartGC()

	return Manager
}

// StartSession returns existing session or creates new session if no matched.
func (m *SessionManager) StartSession(rw http.ResponseWriter, r *http.Request) Session {
	if m == nil {
		log.Panic("session manager does not exists.")
	}

	cookie, err := r.Cookie(m.name)
	agent := r.Header.Get("User-Agent")
	if err != nil || cookie.Value == "" {
		return m.newSession(rw, agent)
	}
	if err != nil {
		log.Error("session manager unable to load cookie [%s]: %s", m.name, err.Error())
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session, _ := m.store.Read(sid)
	if session == nil {
		return m.newSession(rw, agent)
	}

	if token := session.Get("Token"); token != agent {
		return m.newSession(rw, agent)
	}

	if createTime := session.Get("CreateTime"); createTime != nil {
		if createTime.(time.Time).Unix()+int64(m.lapseTime) < time.Now().Unix() {
			// change session id to prevent session hijacking
			newSID := m.newSessionID()
			m.store.UpdateSID(sid, newSID)
			cookie := m.newSessionCookie(newSID)
			http.SetCookie(rw, cookie)
			return session
		}
	}

	return session
}

// EndSession closes session and remove it from session store.
func (m *SessionManager) EndSession(rw http.ResponseWriter, r *http.Request) {
	if m == nil {
		log.Panic("session manager does not exists")
	}

	cookie, err := r.Cookie(m.name)
	if err != nil || cookie.Value == "" {
		return
	}

	if err := m.store.Delete(cookie.Value); err != nil {
		log.Error("unable to delete session from store, error: %s", err.Error())
	}

	http.SetCookie(rw, m.newEndSessionCookie())
}

// StartGC starts the GC for session manager.
func (m *SessionManager) StartGC() {
	if m.isGCStarted {
		return
	}

	m.isGCStarted = true
	m.gcStop = make(chan struct{}, 1)
	m.gc()
}

// StopGC sends signal to stop the GC of session manager.
func (m *SessionManager) StopGC() {
	if !m.isGCStarted {
		return
	}
	m.isGCStarted = false
	close(m.gcStop)
}

// gc runs GC in a new goroutine.
func (m *SessionManager) gc() {
	ticker := time.NewTicker(time.Duration(m.maxLifeTime) * time.Second)
	go func() {
		for {
			select {
			case _, ok := <-m.gcStop:
				if !ok {
					return
				}
			case <-ticker.C:
				m.store.GC(m.maxLifeTime)
			}
		}
	}()
}

// newSessionCookie creates new *http.Cookie object with settings
// defined in session manager.
func (m *SessionManager) newSessionCookie(sid string) *http.Cookie {
	return &http.Cookie{
		Name:     m.name,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   m.maxLifeTime,
	}
}

// newEndSessionCookie creates new empty *http.Cookie object.
// with -1 max age.
func (m *SessionManager) newEndSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     m.name,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}
}

// newSession creates and returns new session object.
func (m *SessionManager) newSession(rw http.ResponseWriter, agent string) Session {
	sid := m.newSessionID()
	session, _ := m.store.Insert(sid, agent)
	cookie := m.newSessionCookie(sid)
	http.SetCookie(rw, cookie)
	return session
}

// newSessionID returns new UUID for session.
func (m *SessionManager) newSessionID() string {
	return uuid.New().String()
}
